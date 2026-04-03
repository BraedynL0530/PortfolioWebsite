from celery import shared_task
from ..models import Repositorys, SummarizerModel
import torch
import joblib
import os
from django.conf import settings


def _load_model(version=None):
    if version is not None:
        record = SummarizerModel.objects.get(version=version)
    else:
        record = SummarizerModel.objects.order_by('-version').first()

    if record is None:
        raise ValueError("No SummarizerModel entries found in the database.")

    full_path = os.path.join(settings.BASE_DIR, record.model_path)

    checkpoint = torch.load(full_path, map_location='cpu')

    from models.model import Seq2SeqModel, config #just the architexure from model in a py file

    cfg = config(
        vocab_size=50260,
        encoder_block_size=1024,
        decoder_block_size=256,
        n_layer=4,
        n_head=8,
        n_embd=320,
        dropout=0.0,
        pad_token_id=50257,
        special_token_ids=[50258, 50259],
    )

    model = Seq2SeqModel(cfg)
    model.load_state_dict(checkpoint['model_state_dict'])
    model.eval()

    # tokenizer assumed to sit next to the model file
    tokenizer_path = full_path.replace('.pt', '_tokenizer.joblib')
    tokenizer = joblib.load(tokenizer_path)

    return model, tokenizer


@shared_task
def summarizeContent(readme, repo_name, version=None):
    DEVICE     = 'cpu'
    SEP_ID     = 50258
    EOS_ID     = 50259

    model, tokenizer = _load_model(version=version)

    # tokenize + truncate
    enc_ids = tokenizer.encode(readme)
    if len(enc_ids) > 1024:
        enc_ids = enc_ids[:1024]

    encoder_input_ids = torch.tensor([enc_ids], dtype=torch.long, device=DEVICE)

    with torch.no_grad():
        output_ids = model.summarize(
            encoder_input_ids,
            max_new_tokens=40,
            temperature=0.7,
            top_k=40,
            bos_token_id=SEP_ID,
            eos_token_id=EOS_ID,
            repetition_penalty=1.5,
        )

    generated = output_ids[0, 1:].tolist()
    if EOS_ID in generated:
        generated = generated[:generated.index(EOS_ID)]

    summary = tokenizer.decode(generated, skip_special_tokens=True)

    Repositorys.objects.filter(name=repo_name).update(summary=summary)