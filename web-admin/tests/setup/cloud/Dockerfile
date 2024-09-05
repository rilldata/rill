FROM python:3.9

WORKDIR /usr/src/app

# copy ui files
COPY web-admin/build .

# start a simple http server using python
ENTRYPOINT ["python3"]
CMD ["-m", "http.server", "3000"]
