terraform {
  required_version = ">= 1.8.0"
  required_providers {
    aws = { source = "hashicorp/aws"; version = "~> 5.0" }
    kubernetes = { source = "hashicorp/kubernetes"; version = "~> 2.0" }
    helm = { source = "hashicorp/helm"; version = "~> 2.0" }
  }
  backend "s3" {
    bucket         = "zu-terraform-state-prod"
    key            = "prod/terraform.tfstate"
    region         = "us-east-1"
    encrypt        = true
    dynamodb_table = "zu-terraform-locks"
  }
}

provider "aws" {
  region = var.aws_region
  default_tags { tags = { Project = "zu-seguridad-360", Environment = "prod", ManagedBy = "terraform" } }
}

module "vpc" {
  source  = "../../modules/vpc"
  env     = "prod"
  cidr    = "10.0.0.0/16"
}

module "eks" {
  source              = "../../modules/eks"
  cluster_name        = "zu-prod"
  private_subnet_ids  = module.vpc.private_subnet_ids
  node_instance_types = ["t3.xlarge"]
  desired_nodes       = 4
  min_nodes           = 3
  max_nodes           = 20
}

module "rds" {
  source          = "../../modules/rds"
  env             = "prod"
  instance_class  = "db.r6g.xlarge"
  multi_az        = true
  subnet_ids      = module.vpc.private_subnet_ids
}

module "redis" {
  source         = "../../modules/redis"
  env            = "prod"
  node_type      = "cache.r6g.large"
  num_cache_nodes = 2
}

module "s3" {
  source = "../../modules/s3"
  env    = "prod"
  worm_retention_years = 20
}

module "vault" {
  source     = "../../modules/vault"
  env        = "prod"
  cluster_id = module.eks.cluster_id
}
