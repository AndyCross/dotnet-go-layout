using System.Net.Http.Headers;
using System.Text;
using System.Text.Json;
using System.Text.Json.Nodes;
using Microsoft.AspNetCore.Mvc.Testing;

namespace Tabanco.Api.Tests;

/// <summary>
/// Runs contract/http.json, the evening at the bar that the Go tests also run,
/// against a fresh in-process server.
/// </summary>
public class ContractTests
{
    // The csproj copies the contract next to the test binary.
    private static readonly string ContractDir = Path.Combine(AppContext.BaseDirectory, "contract");

    [Fact]
    public async Task TheEveningGoesAsTheContractSays()
    {
        await using var factory = new WebApplicationFactory<Program>();
        using var client = factory.CreateClient(new() { AllowAutoRedirect = false });
        var contract = JsonNode.Parse(await File.ReadAllTextAsync(Path.Combine(ContractDir, "http.json"), TestContext.Current.CancellationToken))!;

        var i = 0;
        foreach (var step in contract["steps"]!.AsArray())
        {
            var say = $"step {i++} ({step!["say"]})";
            using var request = new HttpRequestMessage(new HttpMethod((string)step["method"]!), (string)step["path"]!);
            if (step["body"] is { } body)
            {
                request.Content = new StringContent(body.ToJsonString(), Encoding.UTF8, "application/json");
            }
            else if (step["raw"] is { } raw)
            {
                request.Content = new StringContent((string)raw!, Encoding.UTF8, "application/json");
            }

            using var response = await client.SendAsync(request, TestContext.Current.CancellationToken);
            var got = await response.Content.ReadAsStringAsync(TestContext.Current.CancellationToken);

            Assert.True((int)step["status"]! == (int)response.StatusCode, $"{say}: status {(int)response.StatusCode}\n{got}");
            if (step["location"] is { } location)
            {
                Assert.Equal((string)location!, response.Headers.Location?.OriginalString);
            }

            var want = step["jsonFile"] is { } jsonFile
                ? JsonNode.Parse(await File.ReadAllTextAsync(Path.Combine(ContractDir, (string)jsonFile!), TestContext.Current.CancellationToken))
                : step["json"];
            if (step["text"] is { } text)
            {
                Assert.Equal((string)text!, got);
            }
            else if (want is not null)
            {
                Assert.True(JsonNode.DeepEquals(want, JsonNode.Parse(got)), $"{say}: got\n{got}\nwant\n{want.ToJsonString()}");
            }
        }
    }
}
