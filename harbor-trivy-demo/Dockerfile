FROM node:22-alpine

WORKDIR /app

COPY package*.json ./

RUN npm ci --omit=dev

COPY server.js .

EXPOSE 7800

CMD ["node", "server.js"]