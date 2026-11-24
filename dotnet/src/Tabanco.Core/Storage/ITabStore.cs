using System.Diagnostics.CodeAnalysis;

namespace Tabanco.Core.Storage;

/// <summary>Where open tabs live, between requests or between runs.</summary>
public interface ITabStore
{
    bool TryGet(string name, [NotNullWhen(true)] out Tab? tab);

    void Put(Tab tab);

    bool Delete(string name);

    IReadOnlyList<Tab> List();
}
