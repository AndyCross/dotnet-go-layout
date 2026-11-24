using System.Diagnostics.CodeAnalysis;
using System.Text.Json;

namespace Tabanco.Core.Storage;

/// <summary>
/// A store for the CLI: one JSON file, read and written whole on every call,
/// which is plenty for one bar. The file has the same shape as the Go CLI's.
/// </summary>
public sealed class FileTabStore(string path) : ITabStore
{
    public bool TryGet(string name, [NotNullWhen(true)] out Tab? tab) => Load().TryGetValue(name, out tab);

    public void Put(Tab tab)
    {
        var tabs = Load();
        tabs[tab.Name] = tab;
        Save(tabs);
    }

    public bool Delete(string name)
    {
        var tabs = Load();
        if (!tabs.Remove(name))
        {
            return false;
        }
        Save(tabs);
        return true;
    }

    public IReadOnlyList<Tab> List() => [.. Load().Values.OrderBy(t => t.Name, StringComparer.Ordinal)];

    private Dictionary<string, Tab> Load() =>
        File.Exists(path)
            ? JsonSerializer.Deserialize(File.ReadAllText(path), Json.Context.Tabs) ?? []
            : [];

    private void Save(Dictionary<string, Tab> tabs) =>
        File.WriteAllText(path, JsonSerializer.Serialize(tabs, Json.Indented.Tabs) + "\n");
}
