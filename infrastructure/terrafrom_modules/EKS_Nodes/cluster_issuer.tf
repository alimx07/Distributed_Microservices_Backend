resource "kubectl_manifest" "letsencrypt" {
  depends_on = [helm_release.cert_manager]
  yaml_body = yamlencode({
    apiVersion = "cert-manager.io/v1"
    kind       = "ClusterIssuer"
    metadata = {
      name = "letsencrypt-${var.environment}"
    }
    spec = {
      acme = {
        email = var.email
        privateKeySecretRef = {
          name = "letsencrypt-${var.environment}"
        }
        server = "https://acme-v02.api.letsencrypt.org/directory"
        solvers = [
          {
            dns01 = {
              # Cert manager will find configs automatically
              route53 = {}
            }
          },
        ]
      }
    }
  })
}