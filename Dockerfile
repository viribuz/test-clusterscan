# Build stage 
# FROM node:18.14 AS build 
FROM node:18-alpine AS build


WORKDIR /usr/app 

# Copy package.json and package-lock.json 


# RUN npm install -g yarn
COPY package*.json yarn.lock ./ 
COPY package*.json .
# Install npm packages, including sharp --tried replacing with yarn instead of npm
RUN yarn install


# Copy the rest of your application's code 
COPY . . 

# Build your application 
# RUN npm install
# RUN npm run build 

RUN yarn install
RUN npm install chart.js@^4.1.1
RUN yarn build
# Production stage 
#FROM node:18-bullseye 
FROM node:18-alpine 

WORKDIR /usr/app 

# Copy built files from the build stage 
COPY --from=build /usr/app/package*.json /usr/app/yarn.lock ./
COPY --from=build /usr/app/build ./build

#RUN npm ci --force
#RUN npm ci
#RUN echo "nameserver 8.8.8.8" > /etc/resolv.conf && yarn install --production
RUN yarn install --production
RUN yarn global add serve

#--frozen-lockfile

# Expose the port the app runs on 
ENV REACT_APP_API_URL=http://localhost:4000
EXPOSE 3000 

# Start the application 
CMD ["serve", "-s", "build", "-l", "3000"]

