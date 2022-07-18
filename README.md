# kubeconfig-import

Little tool to import a kubeconfig into the local kubeconfig.

It's to avoid these repetitive step when install k3s:

- Read kubeconfig from `$HOME/.kube/config` on `macmini`
- Change all the `default` for the config name (like `hollingsworth`)
- import the cluster, context and user into the local kube config

Example usage:

```bash
ssh macmini.spe.quebec cat .kube/config | kubeconfig-import hollingsworth
```
