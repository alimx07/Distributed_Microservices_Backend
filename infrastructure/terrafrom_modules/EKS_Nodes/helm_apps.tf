resource "helm_release" "nginx_ingress" {
  name             = "ingress-nginx"
  repository       = "https://kubernetes.github.io/ingress-nginx"
  chart            = "ingress-nginx"
  namespace        = "ingress-nginx-ns"
  version          = "4.14.3"
  create_namespace = true
  atomic           = true


  values = [
    yamlencode({
      controller = {
        name                        = "controller"
        enableAnnotationValidations = false
        replicaCount                = 2

        service = {
          enabled = true

          external = {
            enabled = true
          }

          annotations = {
            "service.beta.kubernetes.io/aws-load-balancer-type" = "nlb"
          }

          type = "LoadBalancer"
        }
      }
    })
  ]

  depends_on = [ aws_eks_node_group.this ]
}

resource "helm_release" "cert_manager" {
  name             = "cert-manager"
  repository       = "https://charts.jetstack.io"
  chart            = "cert-manager"
  namespace        = "cert-manager-ns"
  version          = "1.19.3"
  create_namespace = true

  values = [ 
    yamlencode({
      installCRDs = true
      replicaCount = 2
      serviceAccount = {
        annotations = {
          "eks.amazonaws.com/role-arn" = module.cert_manager_irsa.iam_role_arn
        }
      }
      securityContext = {
        fsGroup = 1001
      }
    })
   ]

  depends_on = [ aws_eks_node_group.this ]
}


resource "helm_release" "argo_cd" {
  depends_on       = [helm_release.nginx_ingress,helm_release.cert_manager,kubectl_manifest.letsencrypt]
  name             = "argo-cd"
  repository       = "https://argoproj.github.io/argo-helm"
  chart            = "argo-cd"
  namespace        = "argocd-ns"
  version          = "9.4.0"
  create_namespace = true

  values = [
    yamlencode({
      global = {
        domain = "argocd.${var.environment}.${var.zone_url}"
      }
      server = {
        service = {
          type = "ClusterIP"
          replicas = 2
        }
        ingress = {
          enabled = true
          ingressClassName = "nginx"
          annotations = {
            "cert-manager.io/cluster-issuer" = "letsencrypt-${var.environment}"
          }
          # tls = true
        }
      }
      configs = {
        params = {
          server = {
            insecure = true
          }
        }
        secret = {
          # Secrets
          argocdServerAdminPassword = "var.password"
        }
      }
    })
  ]
}


resource "helm_release" "eso" {
  name             = "external-secrets"
  chart            = "external-secrets"
  namespace        = "external-secrets-ns"
  repository       = "https://charts.external-secrets.io"
  version          = "1.3.2"
  timeout          = 300
  atomic           = true
  create_namespace = true

  values = [
    yamlencode({
    installCRDs = true 
    replicaCount = 2
    serviceAccount = {
      annotations = {
        "eks.amazonaws.com/role-arn" = module.external_secrets_irsa.iam_role_arn
      }
    }
  })
  ]

  depends_on = [ aws_eks_node_group.this ]
}