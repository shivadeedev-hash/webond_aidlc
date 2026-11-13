# syntax=docker/dockerfile:1

# ---- build stage ----
# VITE_* values are compile-time in a Vite app: they are read while `npm run
# build` runs and baked into the bundle. They must therefore be passed as build
# args — setting them as container env vars at runtime has no effect.
FROM node:24-alpine AS build
WORKDIR /app

COPY package.json package-lock.json* ./
RUN npm ci

COPY . .
ARG VITE_DATA_SOURCE=api
ARG VITE_API_URL=http://localhost:8080/api
ARG VITE_SHOW_DEMO_HINT=true
ENV VITE_DATA_SOURCE=$VITE_DATA_SOURCE \
    VITE_API_URL=$VITE_API_URL \
    VITE_SHOW_DEMO_HINT=$VITE_SHOW_DEMO_HINT
RUN npm run build

# ---- runtime stage ----
# Static files behind nginx running as a non-root user (port 8080).
FROM nginxinc/nginx-unprivileged:1.27-alpine AS runtime
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /app/dist /usr/share/nginx/html

EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -q -O /dev/null http://127.0.0.1:8080/ || exit 1
