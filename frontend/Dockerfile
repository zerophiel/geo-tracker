FROM node:20-alpine AS build
WORKDIR /app

COPY package.json package-lock.json ./

# Install dependencies (and vite globally)
RUN npm install
RUN npm install -g vite

COPY . .

RUN npm run build

FROM nginx:alpine
COPY --from=build /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
