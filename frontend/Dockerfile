FROM node:20-alpine AS builder
WORKDIR /app

# Install dependencies first (layer caching)
COPY package*.json ./
RUN npm ci

# Copy source and build
COPY . .
RUN npm run build

# Production (Docker Compose): serve with nginx, proxy to http://core:8080
FROM nginx:alpine AS production
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 3000
CMD ["nginx", "-g", "daemon off;"]

# Cloud Run: nginx with envsubst for dynamic CORE_SERVICE_URL injection at runtime
FROM nginx:alpine AS cloudrun
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx-cloudrun.conf /etc/nginx/templates/nginx-cloudrun.conf
COPY docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod +x /docker-entrypoint.sh
EXPOSE 3000
ENTRYPOINT ["/docker-entrypoint.sh"]
