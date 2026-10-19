variable "cluster_name"        { type = string }
variable "kubernetes_version"  { type = string; default = "1.30" }
variable "private_subnet_ids"  { type = list(string) }
variable "public_access"       { type = bool; default = false }
variable "node_instance_types" { type = list(string); default = ["t3.large"] }
variable "desired_nodes"       { type = number; default = 3 }
variable "min_nodes"           { type = number; default = 2 }
variable "max_nodes"           { type = number; default = 10 }
variable "tags"                { type = map(string); default = {} }
