voici la **documentation d'équipe « comment ajouter un endpoint v2 »**, en prenant comme EXEMPLE le HelloWorld !

# Exemple concret : Ajouter un endpoint v2 (Hello World)

## Rappel de la structure standard

```
pkg/kubernetes_v2/
 ├─ kubernetes.go              // Solution : wiring / Endpoint()
 ├─ handler/
 │     └─ hello.go             // Handler HelloHandler
 └─ service/
       └─ hello.go             // Service HelloService
```


## 1. Implémentation du service (logique métier)

Dans `pkg/kubernetes_v2/service/hello.go` :

```go
package service

type HelloService struct{}

func NewHelloService() *HelloService {
    return &HelloService{}
}

func (s *HelloService) GetHelloWorld() string {
    return "hello world from kubernetes v2!"
}
```


## 2. Implémentation du handler REST

Dans `pkg/kubernetes_v2/handler/hello.go` :

```go
package handler

import (
    "github.com/gin-gonic/gin"
    "gitn.sigma.fr/sigma/paas/api/api/pkg/kubernetes_v2/service"
)
// HelloHandler godoc
// @Summary      Hello world message
// @Description  Retourne un "hello world" pour tester l'API v2 Kubernetes
// @Tags         kubernetes-v2
// @Produce      json
// @Success      200 {object} map[string]string "Hello message"
// @Router       /kubernetes/v2/hello [get]
// @Security     Bearer
func HelloHandler(helloService *service.HelloService) gin.HandlerFunc {
    return func(c *gin.Context) {
        msg := helloService.GetHelloWorld()
        c.JSON(200, gin.H{
            "message": msg,
        })
    }
}
```


## 3. Wiring dans la solution (kubernetes.go)

Dans `pkg/kubernetes_v2/kubernetes.go` :

```go
package kubernetes_v2

import (
    "fmt"
    "github.com/gin-gonic/gin"
    "gitn.sigma.fr/sigma/paas/api/api/pkg/config"
    db "gitn.sigma.fr/sigma/paas/api/api/pkg/db/sqlc/kubernetes"
    kubeclient "gitn.sigma.fr/sigma/paas/api/api/pkg/kubernetes/client"
    hellohandler "gitn.sigma.fr/sigma/paas/api/api/pkg/kubernetes_v2/handler/hello"
    "gitn.sigma.fr/sigma/paas/api/api/pkg/kubernetes_v2/service"
    "k8s.io/klog/v2"
)

type KubernetesSolutionV2 struct {
    client        *kubeclient.Client
    queries       *db.Queries
    cfg           *config.Config
    service_hello *service.HelloService
}

func NewKubernetesSolution(cfg *config.Config, client *kubeclient.Client, queries *db.Queries) (*KubernetesSolutionV2, error) {
    if cfg == nil {
        klog.Errorf("V2: config is required")
        return nil, fmt.Errorf("config is required")
    }
    if client == nil {
        klog.Errorf("V2: kubernetes client is required")
        return nil, fmt.Errorf("kubernetes client is required")
    }
    if queries == nil {
        klog.Errorf("V2: database queries are required")
        return nil, fmt.Errorf("database queries are required")
    }
    helloSvc := service.NewHelloService()
    return &KubernetesSolutionV2{
        client:        client,
        queries:       queries,
        cfg:           cfg,
        service_hello: helloSvc,
    }, nil
}

func (s *KubernetesSolutionV2) Name() string    { return "kubernetes" }
func (s *KubernetesSolutionV2) Version() string { return "v2" }

func (s *KubernetesSolutionV2) Endpoint(rg *gin.RouterGroup) {
    v2 := rg.Group("")
    // Mapping du endpoint v2 hello
    v2.GET("/hello", hellohandler.HelloHandler(s.service_hello))
}
```


## 4. Routing final dans le main

Dans ton fichier main, n’ajoute que :

```go
import kubernetesv2 "gitn.sigma.fr/sigma/paas/api/api/pkg/kubernetes_v2"
// ...
// k8s v2
k8sSolV2, err := kubernetesv2.NewKubernetesSolution(cfg, kubeClient, k8sQueries)
cobra.CheckErr(err)
// ...
ss := []solutions.Solution{
    // ... autres modules ...
    k8sSolV2,
}
```

Le routing est automatique il est gere dans la boucle suivante :

```go
for _, s := range ss {
    s.Endpoint(api.Group(fmt.Sprintf("/%s/%s", s.Name(), s.Version())))
}
```


## 5. Résultat à l’appel

- **Requête GET** sur `/api/kubernetes/v2/hello`
- **Réponse** :

```json
{
  "message": "hello world from kubernetes v2!"
}
```


## Résumé doc à donner à l’équipe

> 1. **Créer le service** dans `service/xxx.go` (toute la logique métier)
> 2. **Créer le handler REST** dans `handler/xxx.go` (pas d'accès métier direct)
> 3. **Instancier le service** dans `KubernetesSolutionV2`
> 4. **Wirer la route/relié handler → service** dans `Endpoint()`
> 5. **Appeler le endpoint** dans le main via la solution v2 (déjà prêt)
>
> ➔ **Ajouter tout nouveau endpoint v2 exactement ainsi, pour assurer la clarté et la testabilité de l’API.**


