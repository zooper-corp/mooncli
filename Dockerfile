FROM gcr.io/distroless/base-debian10
WORKDIR /
COPY /mooncli /mooncli
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/mooncli"]