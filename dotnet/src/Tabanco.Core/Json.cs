using System.Text.Encodings.Web;
using System.Text.Json;

namespace Tabanco.Core;

/// <summary>
/// The one set of JSON options, so the API, the CLI's file and the embedded
/// menu all read and write the same shape as the Go version.
/// </summary>
public static class Json
{
    public static JsonSerializerOptions Options { get; } = Configure(new JsonSerializerOptions(JsonSerializerDefaults.Web));

    /// <summary>
    /// camelCase names, and "Pedro Ximénez" written as it is rather than as
    /// "Pedro Ximénez", which is what the default encoder does.
    /// </summary>
    public static JsonSerializerOptions Configure(JsonSerializerOptions options)
    {
        options.Encoder = JavaScriptEncoder.UnsafeRelaxedJsonEscaping;
        return options;
    }
}
