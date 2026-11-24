using System.Text.Encodings.Web;
using System.Text.Json;
using System.Text.Json.Serialization;

namespace Tabanco.Core;

/// <summary>
/// The one set of JSON options, so the API, the CLI's file and the embedded
/// menu all read and write the same shape as the Go version.
/// </summary>
public static class Json
{
    /// <summary>
    /// Core's own serializers, made by the source generator below rather than
    /// by reflection, so the CLI can be compiled ahead of time.
    /// </summary>
    internal static TabancoJsonContext Context { get; } =
        new(Configure(new JsonSerializerOptions(JsonSerializerDefaults.Web)));

    internal static TabancoJsonContext Indented { get; } =
        new(Configure(new JsonSerializerOptions(JsonSerializerDefaults.Web) { WriteIndented = true }));

    /// <summary>
    /// camelCase names, and "Pedro Ximénez" written as it is rather than as
    /// "Pedro Xim\u00E9nez", which is what the default encoder does.
    /// </summary>
    public static JsonSerializerOptions Configure(JsonSerializerOptions options)
    {
        options.Encoder = JavaScriptEncoder.UnsafeRelaxedJsonEscaping;
        return options;
    }
}

[JsonSerializable(typeof(List<Item>), TypeInfoPropertyName = "Items")]
[JsonSerializable(typeof(Dictionary<string, Tab>), TypeInfoPropertyName = "Tabs")]
internal sealed partial class TabancoJsonContext : JsonSerializerContext;
