FROM gcr.io/distroless/static-debian11:nonroot
ENTRYPOINT ["/baton-vmware-vcenter"]
COPY baton-vmware-vcenter /