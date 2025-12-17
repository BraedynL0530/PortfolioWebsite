from django.db import models

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
