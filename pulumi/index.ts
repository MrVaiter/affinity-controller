import * as k8s from "@pulumi/kubernetes";
import * as kx from "@pulumi/kubernetesx";
import { CustomResourceOptions } from "@pulumi/pulumi";

const serviceAccount = new k8s.core.v1.ServiceAccount("creator", {
    metadata: {
        name: "patcher-sa"
    }
})

const role = new k8s.rbac.v1.ClusterRole("patcher", {
    metadata: {
        name: "patcher-role",
    },
    rules: [{
        apiGroups: ["apps"],
        resources: ["deployments"],
        verbs: ["get", "watch", "list", "update", "delete", "patch", "create"]
    },]
})

const rolebinding = new k8s.rbac.v1.ClusterRoleBinding("bind", {
    metadata: { 
        name: "patcher-rb", 
    },
    roleRef: {
        apiGroup: "rbac.authorization.k8s.io",
        kind: "ClusterRole",
        name: role.metadata.name, 
    },
    subjects: [{
        kind: "ServiceAccount",
        name: serviceAccount.metadata.name,  
        namespace: "default", 
    }],
})

const deployment = new kx.Deployment("fiber-deploy", {
    metadata: { name: "fiber" },
    spec: {
        replicas: 1,
        selector: { matchLabels: { app: "fiber" } },
        template: {
            metadata: { labels: { app: "fiber" } },
            spec: {
                serviceAccount: serviceAccount.metadata.name,
                containers: [{ name: "fiber", image: "dmytromolchanov/decorator-controller-b3010fc694def18c0938838372ff4476:latest" }],
            },
        },
    },
});

const service = deployment.createService({
    ports: {http: 3000}
})

const cr = new k8s.yaml.ConfigFile("decorator", {
    file: "decoratorcontroller.yaml",
    transformations: [
        (obj: any, opts: CustomResourceOptions) => {
            obj.spec.hooks.sync.webhook.url = service.metadata.name.apply(v => `http://${v}:3000/api/v1/deccontrol`)
        },
    ]
})
