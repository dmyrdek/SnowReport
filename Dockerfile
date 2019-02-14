FROM golang:1.11.5

LABEL Douglas Myrdek

ENV APP /~/Desktop/go/src/snowreport
WORKDIR /~/Desktop/go/src/snowreport

ADD . $APP

RUN cd ${APP} && go get -v github.com/jamespearly/loggly && go get -v gopkg.in/robfig/cron.v2
RUN go build 

ENV LOGGLY_TOKEN LogglyTokenHere

ENTRYPOINT ./snowreport
