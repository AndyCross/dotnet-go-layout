# chalk built with go build, on the same base image as the .NET native build.
FROM golang:1.27.3 AS build
WORKDIR /src
COPY go/ ./
RUN CGO_ENABLED=0 go build -o /out/chalk ./cmd/chalk

FROM mcr.microsoft.com/dotnet/runtime-deps:10.0
COPY --from=build /out/chalk /app/chalk
