# haproxy-self-reload

Monitor HAProxy configuration file changes to trigger a reload

*Using docker*

```console
$ docker run -v /some/haproxy.cfg:/etc/haproxy/haproxy.cfg:ro aledbf/haproxy-self-reload:0.2
```

*Creating a pod*

```console
$ kubectl create -f ./pod.yaml
```
