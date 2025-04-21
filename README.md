# 🌐 MultiDC Pod Disruption Budget (MultiDCPDB) Operator

The **Multi DC Pod Disruption Budget (MultiDCPDB) Operator** is a Kubernetes operator that extends the concept of Kubernetes [PodDisruptionBudget (PDB)](https://kubernetes.io/docs/concepts/workloads/pods/disruptions/) to support multi-cloud or multi-datacenter environments.

It ensures high availability of applications across multiple failure domains such as regions, availability zones, cloud providers, or physical datacenters by coordinating disruption budgets beyond the cluster scope.

---

## 🚀 Features

- ✅ Enforces custom disruption policies across multiple Kubernetes clusters.
- 🌍 Cloud-provider agnostic: works across AWS, GCP, Azure, on-prem, etc.
- 🔁 Integrates with native `PodDisruptionBudget` semantics.
- 🔒 Helps ensure high availability during rolling upgrades or node failures.
- 📊 Exposes CRD-based status and metrics for observability.

---

## 📦 Custom Resource Definition (CRD)

The operator introduces a new custom resource: `MultidcPodDisruptionBudget` (`multidcpdb` for short).

```yaml
apiVersion: k8s.tochka.com/v1
kind: MultidcPodDisruptionBudget
metadata:
  name: simple-multidcpoddisruptionbudget-ds
  namespace: simple
  labels:
    app: test-app-ds
spec:
  minAvailable: "3"
  maxUnavailable: "1"
  selector:
    app: "test-app-ds"
```


## 🙌 Contributing
We welcome contributions! Please open issues or submit PRs for:

Bug fixes

New features (e.g., dynamic zone discovery, better metrics)

Documentation improvements

## 📄 License
This project is licensed under the MIT License.

Distributed under MIT License

Copyright (c) 2023-2024 Joint Stock Company Tochka. All rights reserved.

Redistribution and use in source and binary forms, with or without modification, are permitted provided that the following conditions are met:

Redistributions of source code must retain the above copyright notice, this list of conditions and the following disclaimer.

Redistributions in binary form must reproduce the above copyright notice, this list of conditions and the following disclaimer in the documentation and/or other materials provided with the distribution.

Neither the name of Joint Stock Company Tochka nor the names of its contributors may be used to endorse or promote products derived from this software without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

MIT License

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
