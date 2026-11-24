using Tabanco.Api;
using Tabanco.Core;
using Tabanco.Core.Storage;

var builder = WebApplication.CreateBuilder(args);
builder.WebHost.UseUrls($"http://localhost:{Environment.GetEnvironmentVariable("PORT") ?? "8080"}");
builder.Services.ConfigureHttpJsonOptions(o => Json.Configure(o.SerializerOptions));
// A body that isn't JSON throws instead of answering a bare 400 outside
// Development, so the middleware below can give it the same body as Go's.
builder.Services.Configure<RouteHandlerOptions>(o => o.ThrowOnBadRequest = true);
builder.Services.AddSingleton(Menu.Default);
builder.Services.AddSingleton<ITabStore, MemoryTabStore>();

var app = builder.Build();

// Turns the rules' exceptions into answers, as fail() does in the Go version.
app.Use(async (context, next) =>
{
    try
    {
        await next(context);
    }
    catch (BadHttpRequestException)
    {
        await Problems.Write(context, StatusCodes.Status400BadRequest, "the body must be JSON");
    }
    catch (NothingToRepeatException e)
    {
        await Problems.Write(context, StatusCodes.Status409Conflict, e.Message);
    }
    catch (TabRuleException e)
    {
        await Problems.Write(context, StatusCodes.Status400BadRequest, e.Message);
    }
});

app.MapGet("/menu", (Menu menu) => menu.Items);

var tabs = app.MapGroup("/tabs");

tabs.MapGet("/", (Menu menu, ITabStore store) => store.List().Select(t => t.Price(menu)));

tabs.MapPost("/", (OpenTab body, Menu menu, ITabStore store) =>
{
    var name = body.Name?.Trim() ?? "";
    if (name.Length == 0)
    {
        return Problems.Of(StatusCodes.Status400BadRequest, "a tab needs a name");
    }
    lock (Bartender.Lock)
    {
        if (store.TryGet(name, out _))
        {
            return Problems.Of(StatusCodes.Status409Conflict, $"{name} already has a tab open");
        }
        var tab = new Tab(name);
        store.Put(tab);
        return Results.Created($"/tabs/{Uri.EscapeDataString(name)}", tab.Price(menu));
    }
});

tabs.MapGet("/{name}", (string name, Menu menu, ITabStore store) =>
    store.TryGet(name, out var tab) ? Results.Ok(tab.Price(menu)) : Problems.NotFound());

tabs.MapGet("/{name}/chalk", (string name, Menu menu, ITabStore store) =>
    store.TryGet(name, out var tab)
        ? Results.Text(Chalk.Draw(tab.Price(menu)), "text/plain; charset=utf-8")
        : Problems.NotFound());

tabs.MapPost("/{name}/rounds", (string name, Round body, Menu menu, ITabStore store) =>
    Bartender.Change(store, name, tab =>
    {
        tab.Order(menu, body.Items ?? []);
        return tab.Price(menu);
    }));

tabs.MapPost("/{name}/penultima", (string name, Menu menu, ITabStore store) =>
    Bartender.Change(store, name, tab =>
    {
        var poured = tab.Penultima(menu);
        return new { poured = poured.Id, bill = tab.Price(menu) };
    }));

tabs.MapDelete("/{name}", (string name, Menu menu, ITabStore store) =>
{
    lock (Bartender.Lock)
    {
        return store.TryGet(name, out var tab) && store.Delete(name)
            ? Results.Ok(tab.Price(menu))
            : Problems.NotFound();
    }
});

app.Run();

namespace Tabanco.Api
{
    internal sealed record OpenTab(string? Name);

    internal sealed record Round(List<string>? Items);
}
