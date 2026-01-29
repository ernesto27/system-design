FROM oven/bun:1-alpine

WORKDIR /app

COPY package.json .
RUN bun install --production

COPY index.ts .

EXPOSE 80

CMD ["bun", "run", "index.ts"]
