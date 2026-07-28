FROM oven/bun:1.3.13 AS build

WORKDIR /app
COPY tinyschool-ui/package.json tinyschool-ui/bun.lock ./
RUN bun install --no-save
COPY tinyschool-ui/ ./
RUN bun run build

FROM node:22-bookworm-slim

ENV NODE_ENV=production
ENV HOST=0.0.0.0
ENV PORT=3000
WORKDIR /app
COPY --from=build --chown=node:node /app/.output ./.output

USER node
EXPOSE 3000
CMD ["node", ".output/server/index.mjs"]
