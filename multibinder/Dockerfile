FROM ruby:2.3.1-slim

RUN gem install multibinder

VOLUME /opt/multibinder

CMD /usr/local/bundle/bin/multibinder /opt/multibinder/multibinder.sock
