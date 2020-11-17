FROM debian:buster

WORKDIR /tdex-daemon

COPY tdexd /
COPY tdex /

RUN install /tdex /bin

# expose trader and operator interface ports
EXPOSE 9945
EXPOSE 9000

CMD /tdexd

