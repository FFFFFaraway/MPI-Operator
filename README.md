[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/FFFFFaraway/MPI-Operator/blob/main/LICENSE)

# MPI Operator

A big part of this project is based on [MPI Operator in Kubeflow](https://github.com/kubeflow/mpi-operator). This project is a stripped down version written according to my own understanding using [kubebuilder](https://book.kubebuilder.io/).

The MPI Operator makes it easy to run allreduce-style distributed training on Kubernetes. Please check out [this blog post](https://medium.com/kubeflow/introduction-to-kubeflow-mpi-operator-and-industry-adoption-296d5f2e6edc) for an introduction to MPI Operator and its industry adoption.

## Installation

You’ll need a Kubernetes cluster to run against. You can use [KIND](https://sigs.k8s.io/kind) to get a local cluster for testing, or run against a remote cluster. You’ll need [kustomize](https://github.com/kubernetes-sigs/kustomize) installed.
**Note:** Your controller will automatically use the current context in your kubeconfig file (i.e. whatever cluster `kubectl cluster-info` shows).

You can deploy the operator by running the following commands. By default, we will create a namespace 'sw-mpi-operator' and deploy everything in it.

```bash
git clone https://github.com/FFFFFaraway/MPI-Operator
cd mpi-operator
make deploy
```

You can check whether the MPI Job custom resource is installed via:

```bash
kubectl get crd
```

The output should include `mpijobs.batch.test.bdap.com` like the following:

```bash
NAME                                       AGE
...
mpijobs.batch.test.bdap.com                4d
...
```

You can check whether the MPI Job Operator is running via:

```bash
kubectl get pod -n sw-mpi-operator
```

## Creating an MPI Job

You can create an MPI job by defining an `MPIJob` config file. For example:

```yaml
apiVersion: batch.test.bdap.com/v1
kind: MPIJob
metadata:
  name: simple-train
  namespace: sw-mpi-operator
spec:
  numWorkers: 3
  launcherTemplate:
    spec:
      containers:
        - args:
            - sleep 30s && mkdir MPI-Operator && cd MPI-Operator &&
              mkdir sample-python-train && cd sample-python-train &&
              horovodrun -np 2 --hostfile $OMPI_MCA_orte_default_hostfile python main.py
          command:
            - /bin/sh
            - -c
          image: coreharbor.bdap.com/library/horovod-sw-base
          name: horovod-master
      restartPolicy: Never
  workerTemplate:
    spec:
      containers:
        - args:
            - git clone https://github.com/FFFFFaraway/MPI-Operator.git
              && cd sample-python-train && pip install -r requirements.txt && sleep infinity
          command:
            - /bin/sh
            - -c
          image: coreharbor.bdap.com/library/horovod-sw-base
          name: horovod-worker
          resources:
            limits:
              nvidia.com/gpu: 1
      tolerations:
        - effect: NoSchedule
          key: gpu
          operator: Exists
      restartPolicy: OnFailure
```

Deploy the `MPIJob` resource:

```bash
kubectl apply -f config/samples/training_job.yaml
```

## Monitoring an MPI Job

You can inspect the logs to see the training progress. When the job starts, access the logs from the `launcher` pod:

```bash
kubectl logs mpijob-sample-launcher -n sw-mpi-operator
```

## Editing MPI Job

Modify and apply the MPIJob yaml file.

- However, if the Launcher is modified, then you need to manually delete the existing Launcher Pod to trigger the update.
- If the Worker is modified, there is no need to delete Worker Pod manually. It will be automatically updated.

## Deleting MPI Job

Delete the MPIJob yaml file. And all pods, configmaps, rbac will be automatically deleted.Uninstall

UnDeploy the controller to the cluster:

```sh
make undeploy
```

## Docker Images

- [Controller](https://hub.docker.com/r/farawaya/controller)
- [Kubectl-delivery](https://hub.docker.com/r/farawaya/kubectl-delivery)

