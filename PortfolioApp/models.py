from django.db import models
import torch

# Create your models here.
class Repositorys(models.Model):
    name = models.CharField(max_length=100, unique=True)
    readme = models.TextField(blank = True)
    summary = models.TextField(blank = True)
    languages= models.JSONField(default=dict)
    created_at = models.DateTimeField(auto_now_add=True)
    updated_at = models.DateTimeField(auto_now=True)

    class Meta:
        verbose_name_plural = "Repositories"

    def __str__(self):
        return self.name


class SummarizerModel(models.Model):
    name = models.CharField(max_length=100)
    version = models.FloatField()
    model_path = models.CharField(max_length=255)  #models/summarizer_v2.pt or however else

    @property
    def get_model(self):
        return torch.load(self.model_path)
