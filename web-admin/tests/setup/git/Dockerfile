FROM ubuntu

WORKDIR /usr/src/app

RUN apt-get update && apt-get install -y git

ENTRYPOINT ["git"]
CMD ["daemon", "--verbose", "--base-path=/usr/src/app/", "--export-all", "--enable=upload-archive", "--enable=receive-pack", "/usr/src/app/"]
