#!/usr/bin/env python3
import argparse
import csv
import json
from pathlib import Path


def load_mapping(path: Path):
    with path.open('r', encoding='utf-8') as f:
        return json.load(f)


def iter_records(input_path: Path, input_format: str):
    if input_format == 'jsonl':
        with input_path.open('r', encoding='utf-8') as f:
            for line in f:
                line = line.strip()
                if not line:
                    continue
                yield json.loads(line)
    elif input_format == 'json':
        with input_path.open('r', encoding='utf-8') as f:
            data = json.load(f)
        if isinstance(data, list):
            for item in data:
                yield item
        else:
            raise ValueError('json input must be an array')
    elif input_format == 'csv':
        with input_path.open('r', encoding='utf-8', newline='') as f:
            reader = csv.DictReader(f)
            yield from reader
    else:
        raise ValueError(f'unsupported input_format: {input_format}')


def get_by_path(record, path, default=None):
    if not path:
        return default
    cur = record
    for part in path.split('.'):
        if isinstance(cur, dict) and part in cur:
            cur = cur[part]
        else:
            return default
    return cur


def format_prompt(record, mapping):
    template = mapping.get('prompt_template')
    if template:
        flat = flatten_record(record)
        return template.format(**flat)
    prompt_field = mapping.get('field_map', {}).get('prompt')
    if prompt_field:
        return get_by_path(record, prompt_field, '')
    raise ValueError('mapping must define prompt_template or field_map.prompt')


def flatten_record(record, prefix=''):
    out = {}
    if isinstance(record, dict):
        for key, value in record.items():
            name = f'{prefix}{key}' if not prefix else f'{prefix}.{key}'
            if isinstance(value, dict):
                out.update(flatten_record(value, name))
            else:
                out[key] = value
                out[name] = value
    return out


def build_choices(record, mapping):
    items = []
    for spec in mapping.get('choice_fields', []):
        value = get_by_path(record, spec['field'], '')
        label = spec.get('label', '')
        if label:
            items.append(f'{label}. {value}')
        else:
            items.append(str(value))
    return items


def build_sample(record, mapping, index):
    field_map = mapping.get('field_map', {})
    sample = {
        'id': str(get_by_path(record, field_map.get('id'), f'sample-{index}')),
        'benchmark': mapping.get('benchmark', ''),
        'split': mapping.get('split', ''),
        'category': mapping.get('category', ''),
        'prompt': format_prompt(record, mapping),
        'evaluator': mapping.get('evaluator', 'exact_match'),
    }

    for dest, src in field_map.items():
        if dest in ('id', 'prompt'):
            continue
        value = get_by_path(record, src)
        if value not in (None, ''):
            sample[dest] = value

    choices = build_choices(record, mapping)
    if choices:
        sample['choices'] = choices

    acceptable_fields = mapping.get('acceptable_answer_fields', [])
    acceptable_answers = []
    for field in acceptable_fields:
        value = get_by_path(record, field)
        if value not in (None, ''):
            acceptable_answers.append(value)
    if acceptable_answers:
        sample['acceptable_answers'] = acceptable_answers

    required_substrings = mapping.get('required_substrings', [])
    if required_substrings:
        sample['required_substrings'] = required_substrings

    defaults = mapping.get('defaults', {})
    for key, value in defaults.items():
        sample.setdefault(key, value)

    metadata_fields = mapping.get('metadata_fields', {})
    if metadata_fields:
        meta = {}
        for name, field in metadata_fields.items():
            value = get_by_path(record, field)
            if value not in (None, ''):
                meta[name] = value
        if meta:
            sample['metadata'] = meta

    return {k: v for k, v in sample.items() if v not in ('', [], {}, None)}


def main():
    parser = argparse.ArgumentParser(description='Convert raw benchmark/internal datasets into provider-probe JSONL format.')
    parser.add_argument('--mapping', required=True, help='Path to mapping JSON file')
    parser.add_argument('--input', required=True, help='Raw dataset path')
    parser.add_argument('--output', required=True, help='Output JSONL path')
    args = parser.parse_args()

    mapping_path = Path(args.mapping)
    input_path = Path(args.input)
    output_path = Path(args.output)

    mapping = load_mapping(mapping_path)
    input_format = mapping.get('input_format', 'jsonl')

    output_path.parent.mkdir(parents=True, exist_ok=True)
    count = 0
    with output_path.open('w', encoding='utf-8') as out:
        for idx, record in enumerate(iter_records(input_path, input_format), start=1):
            sample = build_sample(record, mapping, idx)
            out.write(json.dumps(sample, ensure_ascii=False) + '\n')
            count += 1
    print(json.dumps({'output': str(output_path), 'samples': count}, ensure_ascii=False))


if __name__ == '__main__':
    main()
