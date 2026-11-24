# chalk compiled ahead of time to one native binary. Native AOT needs clang
# and the zlib headers, and a current SDK: the 10.0.101 SDK in global.json
# picks a native compiler with a known vulnerability (GHSA-p3jp-9v83-76w9).
FROM mcr.microsoft.com/dotnet/sdk:10.0 AS build
RUN apt-get update && apt-get install -y --no-install-recommends clang zlib1g-dev && rm -rf /var/lib/apt/lists/*
WORKDIR /src
COPY contract/ contract/
COPY dotnet/ dotnet/
RUN rm dotnet/global.json \
 && dotnet publish dotnet/src/Tabanco.Cli -c Release -p:PublishAot=true -o /out

FROM mcr.microsoft.com/dotnet/runtime-deps:10.0
COPY --from=build /out/chalk /app/chalk
