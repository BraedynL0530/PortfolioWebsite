from celery import shared_task
from ..models import Repositorys
import torch

@shared_task
def summarizeContent(readme,repo_name):
    summary = torch.load(readme)#placeholder for real model using summirzation
    Repositorys.objects.filter(name=repo_name).update(summary=summary)
    pass