#!/usr/bin/env python3
"""Check bilingual book coverage, structure, links, and executable-code parity.

Run from any directory. --partial checks completed pages during translation;
the default requires both complete editions in the published table of contents.
Go compilation, race tests, and vet: go run scripts/verify-go-examples.go
Mermaid rendering: scripts/verify-mermaid.sh en/part-*/*.md en/appendices/*.md
"""
from collections import Counter
from pathlib import Path
from urllib.parse import unquote, urlsplit
import argparse
import re
import sys

ROOT = Path(__file__).resolve().parent.parent
LINK = re.compile(r'(!?)\[([^\]]*)\]\(([^)]+)\)')
FENCE = re.compile(r'^```([^\n]*)\n(.*?)^```[ \t]*$', re.M | re.S)
HEADING = re.compile(r'^#{2,6}\s+(.+)$', re.M)
SECTION = re.compile(r'^((?:[A-C]|\d+)(?:\.\d+)+)(?:\s|[.：:])')
HAN = re.compile(r'[\u3400-\u9fff]')
# Strings are recognized before comments so URLs and escaped quotes are intact.
COMMENTS = re.compile(r'("(?:\\.|[^"\\])*"|\'(?:\\.|[^\'\\])*\'|`[^`]*`)|//[^\n]*|/\*.*?\*/', re.S)
TOKENS = re.compile(r'"(?:\\.|[^"\\])*"|\'(?:\\.|[^\'\\])*\'|`[^`]*`|\w+|[^\s]', re.S)


def code_tokens(text):
    return TOKENS.findall(COMMENTS.sub(lambda m: m.group(1) or '', text))


def links(text):
    return list(LINK.finditer(FENCE.sub('', text)))


def summary_paths(path):
    return [m.group(3) for m in links(path.read_text()) if not urlsplit(m.group(3)).scheme]


def sections(text):
    return [m.group(1) for heading in HEADING.findall(text) if (m := SECTION.match(heading))]


def table_rows(text):
    return sum(line.lstrip().startswith('|') for line in FENCE.sub('', text).splitlines())


def check(partial=False):
    errors = []
    skipped = []
    checked = 0
    code_count = 0
    diagram_count = 0
    sources = [Path('README.md')]
    sources += [p.relative_to(ROOT) for p in sorted(ROOT.glob('part-*/*.md'))]
    sources += [p.relative_to(ROOT) for p in sorted((ROOT / 'appendices').glob('*.md'))]
    sources += [Path('glossary.md')]
    if len(sources) != 27:
        errors.append(f'Expected introduction, 22 chapters, 3 appendices, glossary; found {len(sources)} pages')

    expected = {p.as_posix() for p in sources}
    expected |= {'en/' + p.as_posix() for p in sources}
    published = summary_paths(ROOT / 'SUMMARY.md')
    if Counter(published) != Counter(expected):
        errors.append('Root SUMMARY.md must list each of the 54 reader pages exactly once')
    if Counter(summary_paths(ROOT / 'en/SUMMARY.md')) != Counter(p.as_posix() for p in sources):
        errors.append('English SUMMARY.md must list its 27 pages exactly once')

    for relative in sources:
        source = ROOT / relative
        translated = ROOT / 'en' / relative
        if not translated.is_file():
            skipped.append(str(relative))
            if not partial:
                errors.append(f'Missing English page: {relative}')
            continue
        zh, en = source.read_text(), translated.read_text()
        checked += 1
        if sections(zh) != sections(en):
            errors.append(f'{relative}: numbered sections differ')
        if len(HEADING.findall(zh)) != len(HEADING.findall(en)):
            errors.append(f'{relative}: heading count differs')
        if table_rows(zh) != table_rows(en):
            errors.append(f'{relative}: table row count differs')
        source_blocks, translated_blocks = FENCE.findall(zh), FENCE.findall(en)
        kinds = lambda blocks: ['text' if kind.strip() in ('', 'text') else kind.strip() for kind, _ in blocks]
        if kinds(source_blocks) != kinds(translated_blocks):
            errors.append(f'{relative}: code/diagram block sequence differs')
        else:
            for i, ((kind, a), (_, b)) in enumerate(zip(source_blocks, translated_blocks), 1):
                kind = kind.strip()
                if kind in ('go', 'json', 'sql', 'bash', 'sh', 'http'):
                    code_count += 1
                    if code_tokens(a) != code_tokens(b):
                        errors.append(f'{relative}: executable tokens differ in {kind} block {i}')
                if kind == 'mermaid':
                    diagram_count += 1
                    if HAN.search(b):
                        errors.append(f'{relative}: untranslated Mermaid text in block {i}')
        # Retain every cited source; repeated references may be consolidated.
        external = lambda text: {m.group(3) for m in links(text) if urlsplit(m.group(3)).scheme in ('https', 'http')}
        missing_citations = external(zh) - external(en)
        if missing_citations:
            errors.append(f'{relative}: missing cited URLs: {sorted(missing_citations)}')
        prose = FENCE.sub('', en)
        prose = '\n'.join(line for line in prose.splitlines() if not line.startswith('[中文]('))
        if HAN.search(prose):
            errors.append(f'{relative}: untranslated Chinese outside code')
        if re.search(r'@@(?:GO|CODE)_|\b(?:TODO_TRANSLATE|TRANSLATION_PENDING)\b', en):
            errors.append(f'{relative}: translation placeholder remains')
        if 'deliveries/' in en or 'WORKFLOW.md' in en:
            errors.append(f'{relative}: internal maintenance link in English edition')
        if not any(m.group(2) == '中文' and (translated.parent / urlsplit(m.group(3)).path).resolve() == source for m in links(en)):
            errors.append(f'{relative}: missing English-to-Chinese switch')
        if not any(m.group(2) == 'English' and (source.parent / urlsplit(m.group(3)).path).resolve() == translated for m in links(zh)):
            errors.append(f'{relative}: missing Chinese-to-English switch')

    pages = [ROOT / p for p in sources] + [ROOT / 'en' / p for p in sources]
    pages += [ROOT / 'SUMMARY.md', ROOT / 'en/SUMMARY.md']
    for page in pages:
        if not page.is_file():
            continue
        text = page.read_text()
        if '交付证据索引' in text or re.search(r'\]\([^)]*(?:deliveries/|WORKFLOW\.md)', text):
            errors.append(f'{page.relative_to(ROOT)}: maintenance material in reader content')
        for m in links(text):
            target = urlsplit(m.group(3))
            if target.scheme or not target.path:
                continue
            resolved = (page.parent / unquote(target.path)).resolve()
            if not resolved.exists():
                if partial and resolved.is_relative_to(ROOT / 'en'):
                    continue
                errors.append(f'{page.relative_to(ROOT)}: missing local link {m.group(3)}')
            if page.is_relative_to(ROOT / 'en') and resolved.suffix == '.md' and not resolved.is_relative_to(ROOT / 'en') and m.group(2) != '中文':
                errors.append(f'{page.relative_to(ROOT)}: English reading link leaves English edition: {m.group(3)}')

    for error in errors:
        print('FAIL', error, file=sys.stderr)
    print(f'Checked {checked}/27 English pages; {code_count} executable blocks; {diagram_count} Mermaid diagrams.')
    if skipped:
        print(f'Pending: {len(skipped)} pages.')
    if errors:
        return 1
    print('PASS bilingual structure, code parity, citations, local links, and language switches.')
    return 0


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--partial', action='store_true', help='Allow English pages still being translated')
    args = parser.parse_args()
    raise SystemExit(check(args.partial))
