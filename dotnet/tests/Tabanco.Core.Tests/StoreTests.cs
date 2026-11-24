using Tabanco.Core.Storage;

namespace Tabanco.Core.Tests;

public sealed class StoreTests : IDisposable
{
    private readonly string _dir = Directory.CreateTempSubdirectory().FullName;

    public void Dispose() => Directory.Delete(_dir, recursive: true);

    public static TheoryData<string> Kinds => ["memory", "file"];

    // Both stores must behave the same, so the same test runs against each.
    [Theory]
    [MemberData(nameof(Kinds))]
    public void BehaveTheSame(string kind)
    {
        ITabStore store = kind == "memory" ? new MemoryTabStore() : new FileTabStore(Path.Combine(_dir, "chalk.json"));

        Assert.False(store.TryGet("Mari", out _));
        store.Put(new Tab("Mari", new() { ["fino"] = 2 }));
        store.Put(new Tab("Antonio"));

        Assert.True(store.TryGet("Mari", out var mari));
        Assert.Equal(2, mari.Marks["fino"]);
        Assert.Equal(["Antonio", "Mari"], store.List().Select(t => t.Name));

        Assert.True(store.Delete("Mari"));
        Assert.False(store.Delete("Mari"));
    }
}
