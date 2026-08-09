FROM node:22-alpine AS builder

WORKDIR /app
COPY frontend/package*.json ./
RUN npm ci

COPY frontend/ .
COPY proto /app/proto

RUN npm run build

FROM node:22-alpine
WORKDIR /app
COPY --from=builder /app/package*.json ./
COPY --from=builder /app/.next ./.next
COPY --from=builder /app/public ./public
COPY --from=builder /app/node_modules ./node_modules
COPY --from=builder /app/proto ./proto

# Default env vars for Next.js to connect to Go services within docker
ENV PRODUCT_SERVICE_URL=product-service:50051
ENV INVENTORY_SERVICE_URL=inventory-service:50052
ENV CART_SERVICE_URL=cart-service:50053

CMD ["npm", "start"]
