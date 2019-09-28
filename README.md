# snmp-trap
Image pushed to registry.npf.dk/snmp-trap, which is hosted locally and such no public docker image is available, but can be build and pushed from the Dockerfile

## Kubernetes caveats 
Due to source IP preservation smnp must be shipped to the node (and nodeport) where the pod is located, this is due to limitation in __externalTrafficPolicy__ where packets that needs to be proxied to reach its destination is dropped. more information can be found here [here](https://kubernetes.io/docs/tutorials/services/source-ip/#source-ip-for-services-with-type-nodeport).
