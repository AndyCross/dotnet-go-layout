using System.Text.Json;

namespace Tabanco.Core;

/// <summary>One thing the bar sells. Price is in euro cents.</summary>
public sealed record Item(string Id, string Name, string Kind, int Price);

/// <summary>The items in the order they are chalked on the board.</summary>
public sealed class Menu
{
    private readonly List<Item> _items;
    private readonly Dictionary<string, Item> _byId;

    public Menu(IEnumerable<Item> items)
    {
        _items = [.. items];
        _byId = _items.ToDictionary(i => i.Id);
    }

    /// <summary>The menu embedded from contract/menu.json, parsed once.</summary>
    public static Menu Default => s_default.Value;

    private static readonly Lazy<Menu> s_default = new(() =>
    {
        using var stream = typeof(Menu).Assembly.GetManifestResourceStream("Tabanco.Core.menu.json")
            ?? throw new InvalidOperationException("menu.json is not embedded");
        var items = JsonSerializer.Deserialize<List<Item>>(stream, Json.Options)
            ?? throw new InvalidOperationException("menu.json is empty");
        return new Menu(items);
    });

    public IReadOnlyList<Item> Items => _items;

    public bool TryFind(string id, out Item item) => _byId.TryGetValue(id, out item!);
}
