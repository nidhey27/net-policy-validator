FROM ubuntu

COPY ./pod-label-validator ./pod-label-validator

ENTRYPOINT [ "./pod-label-validator" ]