# chalk as an ordinary framework-dependent .NET app, run by the JIT.
FROM mcr.microsoft.com/dotnet/sdk:10.0 AS build
WORKDIR /src
COPY contract/ contract/
COPY dotnet/ dotnet/
# The image's SDK is newer than the one global.json names; let it build.
RUN rm dotnet/global.json \
 && dotnet publish dotnet/src/Tabanco.Cli -c Release -o /out

FROM mcr.microsoft.com/dotnet/runtime:10.0
COPY --from=build /out/ /app/
