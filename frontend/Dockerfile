FROM node:22-alpine

WORKDIR /app

COPY package.json ./
RUN npm install

COPY . ./

ARG NEXT_PUBLIC_API_URL=http://localhost:8080
ENV NEXT_PUBLIC_API_URL=$NEXT_PUBLIC_API_URL
ENV HOSTNAME=0.0.0.0

RUN npm run build

EXPOSE 3000
CMD ["npm", "run", "start"]
