FROM node:18-alpine3.17

# RUN apt-get update && apt-get install -y ca-certificates

RUN mkdir /rill-developer
WORKDIR /rill-developer
COPY package*.json .
# RUN npm install .

EXPOSE 3000


ENTRYPOINT ["/bin/sh", "-c" , "npm install -w web-admin && npm run dev -w web-admin"]