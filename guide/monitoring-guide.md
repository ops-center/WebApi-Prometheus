# Monitoring App Using Prometheus and Grafana 

In this guide we are going to describe the process of monitoring the [http-api-server](/example/server.go) using prometheus. We will also use Grafana as visualization tool. 

## Setup Monitoring Environment 

For detailed descriptions visit official documentation of [prometheus-operator](https://github.com/coreos/prometheus-operator/blob/master/Documentation/user-guides/getting-started.md). 

Before deploying prometheus operator, we need to create service account, clusterRole, and clusterRoleBinding to give the operator necessary permissions. Apply the following command to create:

```console
$ kubectl apply -f example/artifacts/prom-rbac.yaml 
  clusterrole.rbac.authorization.k8s.io/prometheus-operator created
  serviceaccount/prometheus-operator created
  clusterrolebinding.rbac.authorization.k8s.io/prometheus-operator created
```

<details>
<summary>
prom-rabc.yaml
</summary>

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  labels:
    app.kubernetes.io/component: controller
    app.kubernetes.io/name: prometheus-operator
    app.kubernetes.io/version: v0.31.1
  name: prometheus-operator
rules:
- apiGroups:
  - apiextensions.k8s.io
  resources:
  - customresourcedefinitions
  verbs:
  - '*'
- apiGroups:
  - monitoring.coreos.com
  resources:
  - alertmanagers
  - prometheuses
  - prometheuses/finalizers
  - alertmanagers/finalizers
  - servicemonitors
  - podmonitors
  - prometheusrules
  verbs:
  - '*'
- apiGroups:
  - apps
  resources:
  - statefulsets
  verbs:
  - '*'
- apiGroups:
  - ""
  resources:
  - configmaps
  - secrets
  verbs:
  - '*'
- apiGroups:
  - ""
  resources:
  - pods
  verbs:
  - list
  - delete
- apiGroups:
  - ""
  resources:
  - services
  - services/finalizers
  - endpoints
  verbs:
  - get
  - create
  - update
  - delete
- apiGroups:
  - ""
  resources:
  - nodes
  verbs:
  - list
  - watch
- apiGroups:
  - ""
  resources:
  - namespaces
  verbs:
  - get
  - list
  - watch

---
apiVersion: v1
kind: ServiceAccount
metadata:
  labels:
    app.kubernetes.io/component: controller
    app.kubernetes.io/name: prometheus-operator
    app.kubernetes.io/version: v0.31.1
  name: prometheus-operator
  namespace: default

---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  labels:
    app.kubernetes.io/component: controller
    app.kubernetes.io/name: prometheus-operator
    app.kubernetes.io/version: v0.31.1
  name: prometheus-operator
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: prometheus-operator
subjects:
- kind: ServiceAccount
  name: prometheus-operator
  namespace: default
```
</details>

Deploy prometheus operator and service for the operator:

```console
$ kubectl apply -f example/artifacts/prom-deploy.yaml 
service/prometheus-operator created
deployment.apps/prometheus-operator created
```

<details>
<summary>prom-deploy.yaml</summary>

```yaml
apiVersion: v1
kind: Service
metadata:
  labels:
    app.kubernetes.io/component: controller
    app.kubernetes.io/name: prometheus-operator
    app.kubernetes.io/version: v0.31.1
  name: prometheus-operator
  namespace: default
spec:
  clusterIP: None
  ports:
  - name: http
    port: 8080
    targetPort: http
  selector:
    app.kubernetes.io/component: controller
    app.kubernetes.io/name: prometheus-operator
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app.kubernetes.io/component: controller
    app.kubernetes.io/name: prometheus-operator
    app.kubernetes.io/version: v0.31.1
  name: prometheus-operator
  namespace: default
spec:
  replicas: 1
  selector:
    matchLabels:
      app.kubernetes.io/component: controller
      app.kubernetes.io/name: prometheus-operator
  template:
    metadata:
      labels:
        app.kubernetes.io/component: controller
        app.kubernetes.io/name: prometheus-operator
        app.kubernetes.io/version: v0.31.1
    spec:
      containers:
      - args:
        - --kubelet-service=kube-system/kubelet
        - --logtostderr=true
        - --config-reloader-image=quay.io/coreos/configmap-reload:v0.0.1
        - --prometheus-config-reloader=quay.io/coreos/prometheus-config-reloader:v0.31.1
        image: quay.io/coreos/prometheus-operator:v0.31.1
        name: prometheus-operator
        ports:
        - containerPort: 8080
          name: http
        resources:
          limits:
            cpu: 200m
            memory: 200Mi
          requests:
            cpu: 100m
            memory: 100Mi
        securityContext:
          allowPrivilegeEscalation: false
      nodeSelector:
        beta.kubernetes.io/os: linux
      securityContext:
        runAsNonRoot: true
        runAsUser: 65534
      serviceAccountName: prometheus-operator
```
</details>

Since we have prometheus operator running, we can deploy prometheus server using it. We also need service account, clusterRole, and clusterRoleBinding for prometheus server.

```console
$ kubectl apply -f example/artifacts/prom-server-rbac.yaml 
clusterrole.rbac.authorization.k8s.io/prometheus created
serviceaccount/prometheus created
clusterrolebinding.rbac.authorization.k8s.io/prometheus created

$ kubectl apply -f example/artifacts/prom-server.yaml 
prometheus.monitoring.coreos.com/prometheus created

$ kubectl apply -f example/artifacts/prom-server-svc.yaml 
service/prometheus created
```

<details>
<summary>prom-server-rbac.yaml</summary>

```yaml
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  name: prometheus
rules:
- apiGroups: [""]
  resources:
  - nodes
  - services
  - endpoints
  - pods
  verbs: ["get", "list", "watch"]
- apiGroups: [""]
  resources:
  - configmaps
  verbs: ["get"]
- nonResourceURLs: ["/metrics"]
  verbs: ["get"]
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: prometheus

---

apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: prometheus
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: prometheus
subjects:
- kind: ServiceAccount
  name: prometheus
  namespace: default
```
</details>

<details>
<summary>prom-server.yaml</summary>

```yaml
apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  name: prometheus
spec:
  serviceAccountName: prometheus
  serviceMonitorSelector:
    matchLabels:
      team: frontend
  resources:
    requests:
      memory: 400Mi
  enableAdminAPI: false
```
</details>

<details>
<summary>prom-server-svc.yaml</summary>

```yaml
apiVersion: v1
kind: Service
metadata:
  name: prometheus
spec:
  type: NodePort
  ports:
  - name: web
    nodePort: 30900
    port: 9090
    protocol: TCP
    targetPort: web
  selector:
    prometheus: prometheus
```

</details>

Check whether they have installed successfully or not:

```console
$ kubectl get pods
NAME                                   READY   STATUS    RESTARTS   AGE
prometheus-operator-69bd579bf9-khqgr   1/1     Running   0          21m
prometheus-prometheus-0                3/3     Running   1          8m59s

```
So, our prometheus server is running and ready to receive metrics. 

Let's install Grafana:

```console
$ kubectl apply -f example/artifacts/grafana.yaml 
deployment.apps/grafana created

$ kubectl get pods -l=app=grafana
NAME                       READY   STATUS    RESTARTS   AGE
grafana-5bd8c6fcf4-l4lzz   1/1     Running   0          2m20s

```
<details>
<summary>grafana.yaml</summary>

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: grafana
  labels:
    app: grafana
spec:
  replicas: 1
  selector:
    matchLabels:
      app: grafana
  template:
    metadata:
      labels:
        app: grafana
    spec:
      containers:
      - name: grafana
        image: grafana/grafana:6.2.5

```

</details>

## Deploy App

Docker image of our [http-web-api](/example/server.go) is available  at `kamolhasan/demoapi:v1alpha1`. You can deploy it by using the following command:

```console
$ kubectl apply -f example/artifacts/demo-app.yaml 
  deployment.apps/demo-server created
```
<details>
<summary>demo-app.yaml</summary>

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: demo-server
  labels:
    app: demo-server
spec:
  replicas: 3
  template:
    metadata:
      name: demo-server
      labels:
        app: demo-server
    spec:
      containers:
        - name: demo-server
          image: kamolhasan/demoapi:v1alpha1
          imagePullPolicy: IfNotPresent
      restartPolicy: Always
  selector:
    matchLabels:
      app: demo-server

```

</details>

Let's deploy a kubernetes service which will communicate with the demo-server's pods.


```console
$ kubectl apply -f example/artifacts/demo-svc.yaml 
  service/demo-server created
```

<details>
<summary>demo-svc.yaml</summary>

```yaml
apiVersion: v1
kind: Service
metadata:
  name: demo-server
  labels:
    app: demo-server
spec:
  selector:
    app: demo-server
  ports:
    - port: 8080
      name: web
  type: NodePort
  
```
</details>

For checking whether we are receiving the prometheus metrics from api-server let's perform the the following commands:

```console
$ kubectl port-forward svc/demo-server 8080
  Forwarding from 127.0.0.1:8080 -> 8080 
  Forwarding from [::1]:8080 -> 8080
  
```
Change the terminal window and perform

```console
$ curl http://localhost:8080/
$ curl -X POST http://localhost:8080/
$ curl http://localhost:8080/metrics
# HELP http_request_duration_seconds HTTP request duration distribution
# TYPE http_request_duration_seconds histogram
http_request_duration_seconds_bucket{method="GET",le="1"} 1
http_request_duration_seconds_bucket{method="GET",le="2"} 1
http_request_duration_seconds_bucket{method="GET",le="5"} 1
http_request_duration_seconds_bucket{method="GET",le="10"} 1
http_request_duration_seconds_bucket{method="GET",le="20"} 1
http_request_duration_seconds_bucket{method="GET",le="60"} 1
http_request_duration_seconds_bucket{method="GET",le="+Inf"} 1
http_request_duration_seconds_sum{method="GET"} 2.11e-07
http_request_duration_seconds_count{method="GET"} 1
http_request_duration_seconds_bucket{method="Post",le="1"} 1
http_request_duration_seconds_bucket{method="Post",le="2"} 1
http_request_duration_seconds_bucket{method="Post",le="5"} 1
http_request_duration_seconds_bucket{method="Post",le="10"} 1
http_request_duration_seconds_bucket{method="Post",le="20"} 1
http_request_duration_seconds_bucket{method="Post",le="60"} 1
http_request_duration_seconds_bucket{method="Post",le="+Inf"} 1
http_request_duration_seconds_sum{method="Post"} 1.13e-07
http_request_duration_seconds_count{method="Post"} 1
# HELP http_requests_total Count of all http requests
# TYPE http_requests_total counter
http_requests_total{code="0",method="GET"} 1
http_requests_total{code="0",method="Post"} 1
# HELP version Version information about this binary
# TYPE version gauge
version{version="v0.0.1"} 0
```

## Monitoring 

Let's consider that you have setup the monitoring environment and deployed an app which can export metrics. If you haven't, complete the previous sections.

Prometheus is needed to be configured to specify a set of targets and parameters describing how to scrape them. Prometheus operator provides [ServiceMonitor](https://github.com/coreos/prometheus-operator/blob/master/Documentation/design.md#servicemonitor) CRD to dynamically configure prometheus by defining how a set of services should be monitored.

```console
$ kubectl apply -f example/artifacts/demo-app-service-monitor.yaml 
  servicemonitor.monitoring.coreos.com/demo-server created
```
<details>
<summary>demo-app-service-monitor.yaml</summary>

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: demo-server
  labels:
    team: frontend
spec:
  selector:
    matchLabels:
      app: demo-server
  endpoints:
    - port: web
```
</details>

Now we need to check whether prometheus has updated its target list. To do so,

```console
$ kubectl get services
  NAME                  TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)          AGE
  demo-server           NodePort    10.107.218.172   <none>        8080:31754/TCP   104m
  kubernetes            ClusterIP   10.96.0.1        <none>        443/TCP          18h
  prometheus            NodePort    10.98.95.103     <none>        9090:30900/TCP   17h
  prometheus-operated   ClusterIP   None             <none>        9090/TCP         17h
  prometheus-operator   ClusterIP   None             <none>        8080/TCP         18h

$ kubectl port-forward svc/prometheus 9090
  Forwarding from 127.0.0.1:9090 -> 9090
  Forwarding from [::1]:9090 -> 9090
 
```
Now go to http://localhost:9090

![home page](/example/images/prom-home.png)

![targets](/example/images/targets.png)

Here we can see, prometheus has updated its target list and source status is `UP`.

Now we can perform query on the metrics we are collecting from our app.

![query-PromQL](/example/images/query.png)

Prometheus dashboard also provides the facility to represent time series data in a graph.

![graph-Prometheus](/example/images/prom-graph.png)

### Grafana 

We have already deployed grafana while setting-up monitoring environment. Let's check again:

```console
$ kubectl get pods -l=app=grafana
  NAME                       READY   STATUS    RESTARTS   AGE
  grafana-5bd8c6fcf4-l4lzz   1/1     Running   1          20h
```
Let's open grafana dashboard by using `kubectl port-forward`. Use username: `admin` and password: `admin` while logging in for the first time.

```console
$ kubectl port-forward grafana-5bd8c6fcf4-l4lzz 3000
  Forwarding from 127.0.0.1:3000 -> 3000
  Forwarding from [::1]:3000 -> 3000
  Handling connection for 3000
```

Visit http://localhost:3000

![Grafana-home](/example/images/grafana-home.png)

Now we need to add data source to grafana dashboard. We need to find the prometheus service endpoint where it is exposing its data.

Our prometheus server exposing metrics at `http://prometheus.default.svc:9090` (format: http://service-name.namespace.svc:port) endpoint.

![grafana-data-source](/example/images/data-source-grafana.png)


Once data source is added and working successfully we can perform query and show the result onto dashboard.

![graph-grafana](/example/images/graph-grafana.png)