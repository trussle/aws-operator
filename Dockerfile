FROM alpine

COPY aws-operator /aws-operator

ENTRYPOINT ["/aws-operator"]
