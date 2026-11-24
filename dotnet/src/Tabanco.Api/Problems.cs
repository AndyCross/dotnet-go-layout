using Microsoft.AspNetCore.WebUtilities;

namespace Tabanco.Api;

/// <summary>
/// Problem details with only title, status and detail, the shape the shared
/// contract pins. TypedResults.Problem would also write a "type" link (it
/// fills one in even when it is set to null), which the Go version would have
/// to invent, so the body is written by hand, as Go's is.
/// </summary>
internal static class Problems
{
    internal sealed record Body(string Title, int Status, string Detail);

    public static IResult Of(int status, string detail) =>
        Results.Json(new Body(ReasonPhrases.GetReasonPhrase(status), status, detail),
            statusCode: status, contentType: "application/problem+json");

    public static IResult NotFound() =>
        Of(StatusCodes.Status404NotFound, "no tab open under that name");

    public static Task Write(HttpContext context, int status, string detail) =>
        Of(status, detail).ExecuteAsync(context);
}
