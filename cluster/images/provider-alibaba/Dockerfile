FROM BASEIMAGE
RUN apk --no-cache add ca-certificates bash

ARG ARCH
ARG TINI_VERSION

ADD provider /usr/local/bin/crossplane-alibaba-provider

COPY stack-package /

EXPOSE 8080
ENTRYPOINT ["crossplane-alibaba-provider"]
