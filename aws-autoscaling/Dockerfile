FROM node:22-alpine

WORKDIR /app

COPY package*.json ./

RUN npm ci --only=production

COPY server.js ./

EXPOSE 3000

USER node

CMD ["node", "server.js"]