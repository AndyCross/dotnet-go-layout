using System.Diagnostics.CodeAnalysis;

namespace Tabanco.Core.Storage;

/// <summary>A store for the API: gone when the process stops.</summary>
public sealed class MemoryTabStore : ITabStore
{
    private readonly Lock _lock = new();
    private readonly Dictionary<string, Tab> _tabs = [];

    public bool TryGet(string name, [NotNullWhen(true)] out Tab? tab)
    {
        lock (_lock)
        {
            tab = _tabs.TryGetValue(name, out var t) ? Clone(t) : null;
            return tab is not null;
        }
    }

    public void Put(Tab tab)
    {
        lock (_lock)
        {
            _tabs[tab.Name] = Clone(tab);
        }
    }

    public bool Delete(string name)
    {
        lock (_lock)
        {
            return _tabs.Remove(name);
        }
    }

    public IReadOnlyList<Tab> List()
    {
        lock (_lock)
        {
            return [.. _tabs.Values.Select(Clone).OrderBy(t => t.Name, StringComparer.Ordinal)];
        }
    }

    private static Tab Clone(Tab t) => new(t.Name, new Dictionary<string, int>(t.Marks));
}
