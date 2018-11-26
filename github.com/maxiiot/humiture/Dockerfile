FROM busybox 

# ENV WORKPATH=/usr/local/humiture/
WORKDIR /usr/local/humiture/
COPY ./ui ./ui
COPY ./app/app .
# COPY ./config/app.toml /etc/humiture/app.toml
ENTRYPOINT [ "./app" ]