import { Construct } from "constructs";
import { App, AssetType, TerraformAsset, TerraformStack } from "cdktf";
import * as google from '@cdktf/provider-google';
import * as path from 'path';

const project = 'web-sakura';
const region = 'us-central1';

class MyStack extends TerraformStack {
  constructor(scope: Construct, id: string) {
    super(scope, id);

    new google.provider.GoogleProvider(this, 'google', {
      project,
    });

    const autoRegistRunner = new google.serviceAccount.ServiceAccount(this, 'autoRegistRunner', {
      accountId: 'auto-regist-runner',
    });

    const accountSecret = new google.secretManagerSecret.SecretManagerSecret(this, 'account', {
      secretId: 'account',
      replication: {
        automatic: true,
      },
    });

    const passwordSecret = new google.secretManagerSecret.SecretManagerSecret(this, 'password', {
      secretId: 'password',
      replication: {
        automatic: true,
      },
    });

    const childIDSecret = new google.secretManagerSecret.SecretManagerSecret(this, 'childID', {
      secretId: 'childID',
      replication: {
        automatic: true,
      },
    });

    const asset = new TerraformAsset(this, 'asset', {
      path: path.resolve('auto-regist'),
      type: AssetType.ARCHIVE,
    });    

    const assetBucket = new google.storageBucket.StorageBucket(this, 'assetBucket', {
      location: region,
      name: `{project}-asset`,
    });

    const assetObject = new google.storageBucketObject.StorageBucketObject(this, 'assetObject', {
      bucket: assetBucket.name,
      name: asset.assetHash,
      source: asset.path,
    });

    new google.cloudfunctions2Function.Cloudfunctions2Function(this, 'autoRegist', {
      buildConfig: {
        entryPoint: 'entry',
        runtime: 'go120',
        source: {
          storageSource: {
            bucket: assetBucket.name,
            object: assetObject.name,
          },
        },
      },
      name: 'auto-regist',
      serviceConfig: {
        minInstanceCount: 0,
        maxInstanceCount: 1,
        secretEnvironmentVariables: [
        {
          key: 'ACCOUNT',
          projectId: project,
          secret: accountSecret.secretId,
          version: 'latest',
        },
        {
          key: 'PASSWORD',
          projectId: project,
          secret: passwordSecret.secretId,
          version: 'latest',
        },
        {
          key: 'CHILD_ID',
          projectId: project,
          secret: childIDSecret.secretId,
          version: 'latest',
        },
        ],
        serviceAccountEmail: autoRegistRunner.email,
      },
    });
  }
}

const app = new App();
new MyStack(app, "web-sakura");
app.synth();
