using Tabanco.Core;
using Tabanco.Core.Storage;

namespace Tabanco.Api;

/// <summary>
/// One bartender: every change to a tab is a read, a change and a write,
/// so they happen one at a time.
/// </summary>
internal static class Bartender
{
    public static readonly Lock Lock = new();

    public static IResult Change<T>(ITabStore store, string name, Func<Tab, T> change)
    {
        lock (Lock)
        {
            if (!store.TryGet(name, out var tab))
            {
                return Problems.NotFound();
            }
            var result = change(tab);
            store.Put(tab);
            return Results.Ok(result);
        }
    }
}
