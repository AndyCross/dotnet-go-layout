using System.Text.Json.Serialization;

namespace Tabanco.Core;

/// <summary>A rule of the bar was broken. The message is for the customer.</summary>
public abstract class TabRuleException(string message) : Exception(message);

public sealed class UnknownItemException(string id) : TabRuleException($"nothing called \"{id}\" on the menu")
{
    public string Id { get; } = id;
}

public sealed class EmptyRoundException() : TabRuleException("a round needs at least one item");

public sealed class NothingToRepeatException() : TabRuleException("nothing to repeat yet: order a drink first");

/// <summary>One item on a bill.</summary>
public sealed record Line(string Item, string Name, int Count, int Cents);

/// <summary>A tab priced against a menu, in board order.</summary>
public sealed record Bill(string Name, IReadOnlyList<Line> Lines, int Total);

/// <summary>What has been chalked on the bar in front of one person.</summary>
public sealed class Tab
{
    [JsonConstructor]
    public Tab(string name, Dictionary<string, int>? marks = null)
    {
        Name = name;
        Marks = marks ?? [];
    }

    public string Name { get; }

    public Dictionary<string, int> Marks { get; }

    /// <summary>
    /// Chalks a round. Every item is checked before any is marked, so a round
    /// with one unknown item marks nothing.
    /// </summary>
    public void Order(Menu menu, params IReadOnlyList<string> ids)
    {
        if (ids.Count == 0)
        {
            throw new EmptyRoundException();
        }
        foreach (var id in ids)
        {
            if (!menu.TryFind(id, out _))
            {
                throw new UnknownItemException(id);
            }
        }
        foreach (var id in ids)
        {
            Marks[id] = Marks.GetValueOrDefault(id) + 1;
        }
    }

    /// <summary>
    /// Orders one more of the drink with the most marks. Nobody in a tabanco
    /// orders the last one, only the one before it. A tie goes to the drink
    /// higher up the board.
    /// </summary>
    public Item Penultima(Menu menu)
    {
        Item? best = null;
        var most = 0;
        foreach (var item in menu.Items)
        {
            var n = Marks.GetValueOrDefault(item.Id);
            if (item.Kind == "drink" && n > most)
            {
                (best, most) = (item, n);
            }
        }
        if (best is null)
        {
            throw new NothingToRepeatException();
        }
        Marks[best.Id]++;
        return best;
    }

    /// <summary>Prices the tab.</summary>
    public Bill Price(Menu menu)
    {
        var lines = menu.Items
            .Where(i => Marks.GetValueOrDefault(i.Id) > 0)
            .Select(i => new Line(i.Id, i.Name, Marks[i.Id], Marks[i.Id] * i.Price))
            .ToList();
        return new Bill(Name, lines, lines.Sum(l => l.Cents));
    }
}
