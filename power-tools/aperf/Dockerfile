FROM public.ecr.aws/aperf/aperf:v1.2.1

RUN dnf install -y findutils && dnf clean all

COPY aperf/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

COPY aperf/send-profile.sh /usr/local/bin/send-profile.sh
RUN chmod +x /usr/local/bin/send-profile.sh

ENTRYPOINT ["/entrypoint.sh"]
