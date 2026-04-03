import torch
import torch.nn as nn
import torch.nn.functional as F
import math
from dataclasses import dataclass, field
from typing import List, Optional

@dataclass
class config:
    vocab_size: int
    encoder_block_size: int
    decoder_block_size: int
    n_layer: int
    n_head: int
    n_embd: int
    dropout: float
    pad_token_id: int
    special_token_ids: List[int] = field(default_factory=list)


class SelfAttention(nn.Module):
    def __init__(self, config: config, is_causal: bool = True, block_size: int = None):
        super().__init__()
        assert config.n_embd % config.n_head == 0
        self.n_head = config.n_head
        self.n_embd = config.n_embd
        self.c_attn = nn.Linear(config.n_embd, 3 * config.n_embd, bias=False)
        self.attn_drop = nn.Dropout(config.dropout)
        self.is_causal = is_causal

        if self.is_causal:
            self.register_buffer(
                "bias",
                torch.tril(torch.ones(block_size, block_size))
                .view(1, 1, block_size, block_size)
            )

        self.c_proj = nn.Linear(config.n_embd, config.n_embd)

    def forward(self, x, pad_mask=None):
        B, T, C = x.size()
        head_dim = C // self.n_head

        qkv = self.c_attn(x)
        q, k, v = qkv.split(C, dim=2)

        q = q.view(B, T, self.n_head, head_dim).transpose(1, 2)
        k = k.view(B, T, self.n_head, head_dim).transpose(1, 2)
        v = v.view(B, T, self.n_head, head_dim).transpose(1, 2)

        att = (q @ k.transpose(-2, -1)) * (1.0 / math.sqrt(head_dim))

        if self.is_causal:
            att = att.masked_fill(self.bias[:, :, :T, :T] == 0, float("-inf"))

        if pad_mask is not None:
            att = att.masked_fill((pad_mask[:, None, None, :T] == 0), float("-inf"))

        att = F.softmax(att, dim=-1)
        att = self.attn_drop(att)

        y = att @ v
        y = y.transpose(1, 2).contiguous().view(B, T, C)
        y = self.c_proj(y)
        return y


class MLP(nn.Module):
    def __init__(self, config: config):
        super().__init__()
        self.fc   = nn.Linear(config.n_embd, 4 * config.n_embd)
        self.proj = nn.Linear(4 * config.n_embd, config.n_embd)
        self.drop = nn.Dropout(config.dropout)

    def forward(self, x):
        return self.drop(self.proj(F.gelu(self.fc(x))))


class Block(nn.Module):
    def __init__(self, config: config, is_causal: bool = True, block_size: int = None):
        super().__init__()
        self.ln_1 = nn.LayerNorm(config.n_embd)
        self.attn = SelfAttention(config, is_causal=is_causal, block_size=block_size)
        self.ln_2 = nn.LayerNorm(config.n_embd)
        self.mlp  = MLP(config)

    def forward(self, x, pad_mask=None):
        x = x + self.attn(self.ln_1(x), pad_mask=pad_mask)
        x = x + self.mlp(self.ln_2(x))
        return x


class CrossAttention(nn.Module):
    def __init__(self, config: config):
        super().__init__()
        assert config.n_embd % config.n_head == 0
        self.n_head = config.n_head
        self.n_embd = config.n_embd

        self.query_proj = nn.Linear(config.n_embd, config.n_embd, bias=False)
        self.key_proj   = nn.Linear(config.n_embd, config.n_embd, bias=False)
        self.value_proj = nn.Linear(config.n_embd, config.n_embd, bias=False)

        self.attn_drop = nn.Dropout(config.dropout)
        self.c_proj    = nn.Linear(config.n_embd, config.n_embd)

    def forward(self, decoder_hidden_states, encoder_output, encoder_pad_mask=None):
        B, T_dec, C = decoder_hidden_states.size()
        _, T_enc, _ = encoder_output.size()
        head_dim = C // self.n_head

        q = self.query_proj(decoder_hidden_states)
        k = self.key_proj(encoder_output)
        v = self.value_proj(encoder_output)

        q = q.view(B, T_dec, self.n_head, head_dim).transpose(1, 2)
        k = k.view(B, T_enc, self.n_head, head_dim).transpose(1, 2)
        v = v.view(B, T_enc, self.n_head, head_dim).transpose(1, 2)

        att = (q @ k.transpose(-2, -1)) * (1.0 / math.sqrt(head_dim))

        if encoder_pad_mask is not None:
            att = att.masked_fill(
                (encoder_pad_mask[:, None, None, :] == 0), float("-inf")
            )

        att = F.softmax(att, dim=-1)
        att = self.attn_drop(att)

        y = att @ v
        y = y.transpose(1, 2).contiguous().view(B, T_dec, C)
        y = self.c_proj(y)
        return y, att


class DecoderBlock(nn.Module):
    def __init__(self, config: config):
        super().__init__()
        self.ln_1      = nn.LayerNorm(config.n_embd)
        self.self_attn = SelfAttention(config, is_causal=True,
                                       block_size=config.decoder_block_size)
        self.ln_2       = nn.LayerNorm(config.n_embd)
        self.cross_attn = CrossAttention(config)
        self.ln_3       = nn.LayerNorm(config.n_embd)
        self.mlp        = MLP(config)

    def forward(self, x, encoder_output, decoder_pad_mask=None, encoder_pad_mask=None):
        x = x + self.self_attn(self.ln_1(x), pad_mask=decoder_pad_mask)
        y, cross_attn_weights = self.cross_attn(
            self.ln_2(x), encoder_output, encoder_pad_mask=encoder_pad_mask
        )
        x = x + y
        x = x + self.mlp(self.ln_3(x))
        return x, cross_attn_weights


class Encoder(nn.Module):
    def __init__(self, config: config):
        super().__init__()
        self.config = config
        self.wte  = nn.Embedding(config.vocab_size, config.n_embd)
        self.wpe  = nn.Embedding(config.encoder_block_size, config.n_embd)
        self.drop = nn.Dropout(config.dropout)
        self.h    = nn.ModuleList([
            Block(config, is_causal=False, block_size=config.encoder_block_size)
            for _ in range(config.n_layer)
        ])
        self.ln_f = nn.LayerNorm(config.n_embd)
        self.apply(self._init_weights)

    def _init_weights(self, module):
        if isinstance(module, nn.Linear):
            nn.init.normal_(module.weight, mean=0.0, std=0.02)
            if module.bias is not None:
                nn.init.zeros_(module.bias)
        elif isinstance(module, nn.Embedding):
            nn.init.normal_(module.weight, mean=0.0, std=0.02)

    def forward(self, input_ids, pad_mask=None):
        B, T = input_ids.size()
        assert T <= self.config.encoder_block_size

        pos     = torch.arange(T, device=input_ids.device).unsqueeze(0)
        x       = self.drop(self.wte(input_ids) + self.wpe(pos))
        for block in self.h:
            x = block(x, pad_mask=pad_mask)
        return self.ln_f(x)


class Decoder(nn.Module):
    def __init__(self, config: config):
        super().__init__()
        self.config = config
        self.wte  = nn.Embedding(config.vocab_size, config.n_embd)
        self.wpe  = nn.Embedding(config.decoder_block_size, config.n_embd)
        self.drop = nn.Dropout(config.dropout)
        self.h    = nn.ModuleList([DecoderBlock(config) for _ in range(config.n_layer)])
        self.ln_f = nn.LayerNorm(config.n_embd)
        self.lm_head = nn.Linear(config.n_embd, config.vocab_size, bias=False)
        self.lm_head.weight = self.wte.weight          # weight tying
        self.p_gen = nn.Linear(config.n_embd, 1)
        self.apply(self._init_weights)

    def _init_weights(self, module):
        if isinstance(module, nn.Linear):
            nn.init.normal_(module.weight, mean=0.0, std=0.02)
            if module.bias is not None:
                nn.init.zeros_(module.bias)
        elif isinstance(module, nn.Embedding):
            nn.init.normal_(module.weight, mean=0.0, std=0.02)

    def forward(self, decoder_input_ids, encoder_output,
                decoder_pad_mask=None, encoder_pad_mask=None):
        B, T = decoder_input_ids.size()
        assert T <= self.config.decoder_block_size

        pos = torch.arange(T, device=decoder_input_ids.device).unsqueeze(0)
        x   = self.drop(self.wte(decoder_input_ids) + self.wpe(pos))

        attn_weight_list = []
        for block in self.h:
            x, attn = block(x, encoder_output,
                            decoder_pad_mask=decoder_pad_mask,
                            encoder_pad_mask=encoder_pad_mask)
            attn_weight_list.append(attn)

        x = self.ln_f(x)

        p_gen       = torch.sigmoid(self.p_gen(x))                        # (B, T, 1)
        final_attn  = sum(attn_weight_list) / len(attn_weight_list)        # avg over layers
        final_attn  = final_attn.mean(dim=1)                               # avg over heads → (B, T_dec, T_enc)
        logits      = self.lm_head(x)

        return logits, final_attn, p_gen


class Seq2SeqModel(nn.Module):

    def __init__(self, config: config):
        super().__init__()
        self.config       = config
        self.pad_token_id = config.pad_token_id
        #build a set of token ids that must never appear in the copy distribution
        self._no_copy_ids = set(config.special_token_ids) | {config.pad_token_id}

        self.encoder = Encoder(config)
        self.decoder = Decoder(config)
        self.apply(self._init_weights)

    def _init_weights(self, module):
        if isinstance(module, nn.Linear):
            nn.init.normal_(module.weight, mean=0.0, std=0.02)
            if module.bias is not None:
                nn.init.zeros_(module.bias)
        elif isinstance(module, nn.Embedding):
            nn.init.normal_(module.weight, mean=0.0, std=0.02)

    def _compute_copy_logits(self, attn_weights, encoder_input_ids):
        B, T_dec, T_enc = attn_weights.shape
        copy_dist = torch.zeros(B, T_dec, self.config.vocab_size,
                                device=attn_weights.device, dtype=attn_weights.dtype)

        # Build mask: True = token is allowed to be copied
        copy_mask = torch.ones_like(encoder_input_ids, dtype=torch.bool)
        for sid in self._no_copy_ids:
            copy_mask &= (encoder_input_ids != sid)

        for b in range(B):
            valid_idx      = torch.where(copy_mask[b])[0]
            if valid_idx.numel() == 0:
                continue
            valid_token_ids = encoder_input_ids[b, valid_idx]
            valid_attn      = attn_weights[b, :, valid_idx]

            index = valid_token_ids.unsqueeze(0).expand(T_dec, -1)
            copy_dist[b].scatter_add_(1, index, valid_attn)

        #renormalise so copy side always sums to 1
        copy_sum = copy_dist.sum(dim=-1, keepdim=True).clamp(min=1e-10)
        copy_dist = copy_dist / copy_sum
        return copy_dist

    def forward(self, encoder_input_ids, decoder_input_ids,
                targets=None, use_scheduled_sampling=False, sampling_prob=0.0):

        encoder_pad_mask = (encoder_input_ids != self.pad_token_id).float()
        encoder_output   = self.encoder(encoder_input_ids, pad_mask=encoder_pad_mask)

        # pass encoder_input_ids so scheduled sampling can use the full
        #pointer-network probability (not just generation logits)
        if use_scheduled_sampling and self.training and sampling_prob > 0:
            decoder_input_ids = self._scheduled_sampling(
                decoder_input_ids, encoder_output,
                encoder_pad_mask, encoder_input_ids, sampling_prob
            )

        decoder_pad_mask = (decoder_input_ids != self.pad_token_id).float()
        logits, attn_weights, p_gen = self.decoder(
            decoder_input_ids, encoder_output,
            decoder_pad_mask=decoder_pad_mask,
            encoder_pad_mask=encoder_pad_mask
        )

        gen_probs   = F.softmax(logits, dim=-1)
        copy_probs  = self._compute_copy_logits(attn_weights, encoder_input_ids)
        final_probs = p_gen * gen_probs + (1 - p_gen) * copy_probs
        final_log_probs = torch.log(final_probs + 1e-10)

        nll_loss            = torch.tensor(0.0, device=logits.device)
        attn_entropy_val    = 0.0
        mean_cross_attn_val = 0.0
        loss                = None

        if targets is not None:
            nll_loss = F.nll_loss(
                final_log_probs.view(-1, self.config.vocab_size),
                targets.view(-1),
                ignore_index=-100,
                reduction='mean'
            )

            valid_attn = attn_weights * encoder_pad_mask.unsqueeze(1)

            attn_entropy = -torch.sum(
                valid_attn * torch.log(valid_attn + 1e-10), dim=-1
            )
            attn_entropy_val = attn_entropy.mean().item()

            max_acceptable_entropy = 3.0
            attention_penalty = F.relu(attn_entropy.mean() - max_acceptable_entropy)

            mean_cross_attn     = valid_attn.sum(dim=-1).mean()
            mean_cross_attn_val = mean_cross_attn.item()

            min_attn_threshold  = 0.5
            attention_usage_penalty = F.relu(min_attn_threshold - mean_cross_attn)

            # Vocabulary diversity penalty
            with torch.no_grad():
                pred_tokens = final_probs.argmax(dim=-1)

            vocab_diversity_losses = []
            for b in range(pred_tokens.size(0)):
                seq       = pred_tokens[b]
                valid_seq = seq[targets[b] != -100]
                if len(valid_seq) > 1:
                    unique_ratio  = len(torch.unique(valid_seq)) / len(valid_seq)
                    diversity_loss = (0.5 - unique_ratio) ** 2
                    vocab_diversity_losses.append(diversity_loss)

            if vocab_diversity_losses:
                vocab_diversity_penalty = torch.tensor(
                    vocab_diversity_losses, device=logits.device
                ).mean()
            else:
                vocab_diversity_penalty = torch.tensor(0.0, device=logits.device)

            gen_encouragement = torch.mean((1 - p_gen) ** 2)

            # shift cumsum by 1 along the decoder dimension so coverage_t = sum_{s<t} attn_s
            cumulative      = torch.cumsum(attn_weights, dim=1)
            prev_coverage   = F.pad(cumulative[:, :-1, :], (0, 0, 1, 0))   # shift right by 1
            coverage_loss   = torch.sum(
                torch.min(attn_weights, prev_coverage), dim=-1
            ).mean()

            loss = (
                nll_loss
                + 0.3  * attention_penalty
                + 0.2  * attention_usage_penalty
                + 0.1  * vocab_diversity_penalty
                + 0.15 * gen_encouragement #updated for based of of v0.1 of current model it copys a bit too much
                + 0.1  * coverage_loss
            )

        return final_log_probs, loss, {
            'nll_loss':        nll_loss.item(),
            'attn_entropy':    attn_entropy_val,
            'mean_cross_attn': mean_cross_attn_val,
            'p_gen_mean':      p_gen.mean().item(),
        }

    # ── Scheduled sampling ────────────────────────────────────────────────────
    def _scheduled_sampling(self, decoder_input_ids, encoder_output,
                            encoder_pad_mask, encoder_input_ids, sampling_prob):
        B, T          = decoder_input_ids.size()
        new_dec_input = decoder_input_ids.clone()

        with torch.no_grad():
            for t in range(1, T):
                if torch.rand(1).item() >= sampling_prob:
                    continue

                partial_input    = new_dec_input[:, :t]
                decoder_pad_mask = torch.ones(
                    (B, t), dtype=torch.float, device=decoder_input_ids.device
                )

                logits, attn_weights, p_gen = self.decoder(
                    partial_input, encoder_output,
                    decoder_pad_mask=decoder_pad_mask,
                    encoder_pad_mask=encoder_pad_mask
                )

                # Use only the last decoder position
                last_gen  = F.softmax(logits[:, -1:, :], dim=-1)
                last_attn = attn_weights[:, -1:, :]
                last_p    = p_gen[:, -1:, :]

                last_copy  = self._compute_copy_logits(last_attn, encoder_input_ids)
                last_probs = (last_p * last_gen + (1 - last_p) * last_copy)[:, 0, :]

                last_probs = last_probs.clamp(min=0)
                last_probs = last_probs / last_probs.sum(dim=-1, keepdim=True).clamp(min=1e-10)

                sampled = torch.multinomial(last_probs, num_samples=1)
                new_dec_input[:, t] = sampled.squeeze(-1)

        return new_dec_input

    @torch.no_grad()
    def summarize(self, encoder_input_ids, max_new_tokens,temperature=1.0, top_k=50,bos_token_id=None, eos_token_id=None,repetition_penalty=1.2):
        if bos_token_id is None:
            raise ValueError("bos_token_id must be provided")

        self.eval()
        encoder_pad_mask = (encoder_input_ids != self.pad_token_id).float()
        encoder_output   = self.encoder(encoder_input_ids, pad_mask=encoder_pad_mask)

        decoder_input_ids = torch.tensor([[bos_token_id]], device=encoder_input_ids.device)

        for _ in range(max_new_tokens):
            cur_dec = decoder_input_ids
            if cur_dec.size(1) > self.config.decoder_block_size:
                cur_dec = cur_dec[:, -self.config.decoder_block_size:]

            B, t = cur_dec.size()
            dec_pad_mask = torch.ones((B, t), dtype=torch.float,
                                      device=encoder_input_ids.device)

            logits, attn_weights, p_gen = self.decoder(
                cur_dec, encoder_output,
                decoder_pad_mask=dec_pad_mask,
                encoder_pad_mask=encoder_pad_mask
            )

            gen_probs  = F.softmax(logits, dim=-1)
            copy_probs = self._compute_copy_logits(attn_weights, encoder_input_ids)
            final_probs = p_gen * gen_probs + (1 - p_gen) * copy_probs   # (B, T, V)

            #convert to log space, apply temperature, re-softmax
            last_log_probs = torch.log(final_probs[:, -1, :].clamp(min=1e-10))

            # Repetition penalty
            if repetition_penalty != 1.0 and decoder_input_ids.size(1) > 1:
                for tok_id in decoder_input_ids[0].tolist():
                    last_log_probs[:, tok_id] /= repetition_penalty

            last_log_probs = last_log_probs / temperature    # temperature scaling


            if top_k > 0:
                topk_vals, _ = torch.topk(last_log_probs, top_k)
                last_log_probs[last_log_probs < topk_vals[:, [-1]]] = float('-inf')

            probs = F.softmax(last_log_probs, dim=-1)

            next_token = torch.multinomial(probs, num_samples=1)
            decoder_input_ids = torch.cat([decoder_input_ids, next_token], dim=1)

            if eos_token_id is not None and next_token.item() == eos_token_id:
                break

        return decoder_input_ids


print("Fixed model loaded")
