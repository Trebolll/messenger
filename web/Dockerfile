FROM node:22-alpine AS build
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
ARG VITE_API_BASE=http://localhost:8080
ARG VITE_WS_BASE=ws://localhost:8080
ENV VITE_API_BASE=$VITE_API_BASE
ENV VITE_WS_BASE=$VITE_WS_BASE
RUN npm run build

FROM nginx:1.27-alpine
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /app/dist /usr/share/nginx/html
EXPOSE 80
