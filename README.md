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
  name: simple-train-cpu
  namespace: sw-mpi-operator
spec:
  numWorkers: 5
  launcherTemplate:
    spec:
      containers:
        - args:
            - mkdir MPI-Operator &&
              cd MPI-Operator &&
              mkdir sample-python-train &&
              cd sample-python-train &&
              horovodrun -np 2 --hostfile $OMPI_MCA_orte_default_hostfile python main.py
          command:
            - /bin/sh
            - -c
          image: farawaya/horovod-torch-cpu
          name: horovod-master
      restartPolicy: Never
  workerTemplate:
    spec:
      containers:
        - args:
            - git clone https://github.com/FFFFFaraway/MPI-Operator.git &&
              cd MPI-Operator &&
              cd sample-python-train &&
              pip install -r requirements.txt &&
              touch /ready.txt &&
              sleep infinity
          command:
            - /bin/sh
            - -c
          image: farawaya/horovod-torch-cpu
          name: horovod-worker
          readinessProbe:
            exec:
              command:
                - cat
                - /ready.txt
            initialDelaySeconds: 30
            periodSeconds: 5
```

Deploy the `MPIJob` resource:

```bash
kubectl apply -f config/samples/training_job_cpu.yaml
```

Note that the launcher pod will use all workers (numWorkers in spec), the `-np`parameter after horovodrun does not seem to work.

## Monitoring an MPI Job

You can inspect the logs to see the training progress. When the job starts, access the logs from the `launcher` pod:

```bash
kubectl logs simple-train-cpu-launcher -n sw-mpi-operator
```

## Editing MPI Job

Modify and apply the MPIJob yaml file.

- However, if the Launcher is modified, then you need to manually delete the existing Launcher Pod to trigger the update.
- If the Worker is modified, there is no need to delete Worker Pod manually. It will be automatically updated.

## Deleting MPI Job

Delete the MPIJob yaml file. And all pods, configmaps, rbac will be automatically deleted.

You need to **manually** delete the MPIJob task to avoid occupying GPU resources.

## Uninstall

```sh
make undeploy
```

## TODO List

- Add MPIJob Status
- Add Defaulter and Validator Webhook

## Docker Images

- [Controller](https://hub.docker.com/r/farawaya/controller)
- [Kubectl-delivery](https://hub.docker.com/r/farawaya/kubectl-delivery)
- [Horovod-torch-cuda113](https://hub.docker.com/r/farawaya/horovod-torch-cuda113)
- [Horovod-torch-cpu](https://hub.docker.com/r/farawaya/horovod-torch-cpu)

