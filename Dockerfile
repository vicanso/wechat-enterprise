FROM node:8-alpine
EXPOSE 3011 
ADD ./ /app
RUN cd /app \
  && npm i --production \
  && sh ./clear.sh
CMD ["node", "/app/app"]