import argparse
import logging
import random
import re
import time

from k8sclient.client import api_client
from k8sclient.client.apis import apiv_api

def get_client(kube_apiserver=None, key_file=None, cert_file=None,
               ca_certs=None):
    kube_apiserver = kube_apiserver
    key_file = key_file
    cert_file = cert_file
    ca_certs = ca_certs
    return api_client.ApiClient(host=kube_apiserver, key_file=key_file,
                                cert_file=cert_file, ca_certs=ca_certs)

def get_v1_api(client):
    return apiv_api.ApivApi(client)

def get_pods(api=None, namespace=None):
    return api.list_namespaced_pod(namespace=namespace).items

def delete_pod(api=None, namespace=None, name=None):
    return api.delete_namespaced_pod({}, namespace=namespace, name=name)

def _get_name(k8s_object):
    return k8s_object.metadata.name

def _get_names(k8s_objects):
    names = []
    for item in k8s_objects or []:
      names.append(_get_name(item))
    return names

def _pick_random_item(item_list):
    return random.choice(item_list)

def pick_and_delete(api, namespace=None, period=None, regexp=None):
    all_pods = _get_names(get_pods(api, namespace=namespace))
    if regexp:
        pick_list = list(filter(regexp.match, all_pods))
    else:
        pick_list = all_pods
    pod_to_delete = _pick_random_item(pick_list)
    delete_pod(api, namespace=namespace, name=pod_to_delete)
    logging.warning("Pod %s deleted" % pod_to_delete)

if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("kube_apiserver",
                        type=str,
                        default='localhost:8080',
                        nargs='?')
    parser.add_argument("namespace",
                        type=str,
                        default="default",
                        nargs='?')
    parser.add_argument("regexp", default=None, nargs='?')
    parser.add_argument("period",
                        type=int,
                        default=10,
                        nargs='?')
    args = parser.parse_args()
    kube_apiserver = args.kube_apiserver
    namespace = args.namespace
    period = args.period
    if args.regexp:
        regexp = args.regexp
        r = re.compile(regexp, re.IGNORECASE)
    else:
        r = None
    client = get_client(kube_apiserver=kube_apiserver)
    api = get_v1_api(client)
    while True:
        pick_and_delete(api=api, namespace=namespace, period=period, regexp = r)
        time.sleep(period)

