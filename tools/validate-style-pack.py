#!/usr/bin/env python3
"""Validate a dumps Style Pack v1 using only the Python standard library."""

import json
import re
import sys
from pathlib import Path


SCHEMA_URL = "https://grippado.github.io/dumps/style-packs/schema/v1.json"
DTCG_SCHEMA_URL = "https://www.designtokens.org/schemas/2025.10/format.json"
FORMATS = {"report", "story", "slides", "dashboard", "doc", "comparison", "changelog", "landing"}
LEVELS = {"classic", "refined", "bold", "avant-garde"}
LAYOUTS = {"column", "editorial", "split", "poster"}
HEX_COLOR = re.compile(r"^#[0-9a-fA-F]{6}$")
FONT_FAMILY = re.compile(r"^[A-Za-z0-9 ,\"'-]+$")
ID = re.compile(r"^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$")


class Validation:
    def __init__(self):
        self.errors = []

    def error(self, path, message):
        self.errors.append(f"{path}: {message}")

    def require_object(self, value, path):
        if not isinstance(value, dict):
            self.error(path, "must be an object")
            return False
        return True

    def require_string(self, value, path):
        if not isinstance(value, str) or not value.strip():
            self.error(path, "must be a non-empty string")
            return False
        return True

    def localized(self, value, path):
        if not self.require_object(value, path):
            return
        for locale in ("en", "pt-BR"):
            self.require_string(value.get(locale), f"{path}.{locale}")

    def token_tree(self, value, path):
        if not self.require_object(value, path):
            return 0
        if "$value" in value:
            if not self.require_string(value.get("$type"), f"{path}.$type"):
                return 1
            if "$value" not in value:
                self.error(path, "token must have $value")
            return 1
        count = 0
        for key, child in value.items():
            if key.startswith("$"):
                continue
            count += self.token_tree(child, f"{path}.{key}")
        if count == 0:
            self.error(path, "group must contain at least one typed token")
        return count

    @staticmethod
    def token_value(group, key):
        token = group.get(key, {}) if isinstance(group, dict) else {}
        if not isinstance(token, dict):
            return None
        return token.get("$value")

    @staticmethod
    def luminance(hex_color):
        rgb = [int(hex_color[i : i + 2], 16) / 255 for i in (1, 3, 5)]
        linear = [channel / 12.92 if channel <= 0.04045 else ((channel + 0.055) / 1.055) ** 2.4 for channel in rgb]
        return 0.2126 * linear[0] + 0.7152 * linear[1] + 0.0722 * linear[2]

    def contrast(self, foreground, background, path):
        if not isinstance(foreground, str) or not HEX_COLOR.fullmatch(foreground):
            self.error(path, "must be a six-digit hex color")
            return
        if not isinstance(background, str) or not HEX_COLOR.fullmatch(background):
            self.error(path, "background must be a six-digit hex color")
            return
        first, second = sorted((self.luminance(foreground), self.luminance(background)), reverse=True)
        ratio = (first + 0.05) / (second + 0.05)
        if ratio < 4.5:
            self.error(path, f"text contrast is {ratio:.2f}:1; needs at least 4.5:1")

    def validate(self, pack):
        if not self.require_object(pack, "$pack"):
            return
        allowed = {"$schema", "schemaVersion", "id", "metadata", "supports", "designTokens", "voice", "guidance", "accessibility", "examples", "assets"}
        for key in pack.keys() - allowed:
            self.error(key, "unknown top-level property")
        if pack.get("$schema") != SCHEMA_URL:
            self.error("$schema", f"must be {SCHEMA_URL}")
        if pack.get("schemaVersion") != 1:
            self.error("schemaVersion", "must be 1")
        if not isinstance(pack.get("id"), str) or not ID.fullmatch(pack["id"]):
            self.error("id", "must be a lowercase slug of up to 63 characters")

        metadata = pack.get("metadata")
        if self.require_object(metadata, "metadata"):
            self.localized(metadata.get("name"), "metadata.name")
            self.localized(metadata.get("summary"), "metadata.summary")
            self.require_string(metadata.get("author"), "metadata.author")
            self.require_string(metadata.get("license"), "metadata.license")

        supports = pack.get("supports")
        if not isinstance(supports, list) or not supports:
            self.error("supports", "must list at least one supported format")
            supports = []
        else:
            seen = []
            for index, format_name in enumerate(supports):
                if not isinstance(format_name, str):
                    self.error(f"supports[{index}]", "must be a string")
                    continue
                if format_name in seen:
                    self.error("supports", "must not contain duplicates")
                seen.append(format_name)
        for index, format_name in enumerate(supports):
            if isinstance(format_name, str) and format_name not in FORMATS:
                self.error(f"supports[{index}]", f"unknown format {format_name!r}")

        tokens = pack.get("designTokens")
        if self.require_object(tokens, "designTokens"):
            if tokens.get("$schema") != DTCG_SCHEMA_URL:
                self.error("designTokens.$schema", f"must be {DTCG_SCHEMA_URL}")
            token_count = 0
            for key, group in tokens.items():
                if key == "$schema":
                    continue
                token_count += self.token_tree(group, f"designTokens.{key}")
            if token_count == 0:
                self.error("designTokens", "must contain typed design tokens")

            colors = tokens.get("color", {})
            fonts = tokens.get("fontFamily", {})
            for scheme in ("light", "dark"):
                palette = colors.get(scheme, {}) if isinstance(colors, dict) else {}
                required_colors = ("page", "ink", "accent", "muted")
                values = {key: self.token_value(palette, key) for key in required_colors}
                for key, color in values.items():
                    if not isinstance(color, str) or not HEX_COLOR.fullmatch(color):
                        self.error(f"designTokens.color.{scheme}.{key}", "must be a six-digit hex color token")
                if all(isinstance(values[key], str) and HEX_COLOR.fullmatch(values[key]) for key in ("page", "ink", "accent", "muted")):
                    for key in ("ink", "accent", "muted"):
                        self.contrast(values[key], values["page"], f"designTokens.color.{scheme}.{key}")
            for key in ("display", "body"):
                family = self.token_value(fonts, key)
                if isinstance(family, list):
                    valid = bool(family) and all(isinstance(item, str) and FONT_FAMILY.fullmatch(item) for item in family)
                else:
                    valid = isinstance(family, str) and bool(FONT_FAMILY.fullmatch(family))
                if not valid:
                    self.error(f"designTokens.fontFamily.{key}", "must be a safe font-family token")

        voice = pack.get("voice")
        if self.require_object(voice, "voice"):
            self.require_string(voice.get("summary"), "voice.summary")
            principles = voice.get("principles")
            if not isinstance(principles, list) or not principles:
                self.error("voice.principles", "must contain at least one principle")
            else:
                for index, principle in enumerate(principles):
                    self.require_string(principle, f"voice.principles[{index}]")
            examples = voice.get("examples")
            if not isinstance(examples, list) or not examples:
                self.error("voice.examples", "must contain at least one before/after example")
            else:
                for index, example in enumerate(examples):
                    if self.require_object(example, f"voice.examples[{index}]"):
                        self.require_string(example.get("before"), f"voice.examples[{index}].before")
                        self.require_string(example.get("after"), f"voice.examples[{index}].after")

        guidance = pack.get("guidance")
        if self.require_object(guidance, "guidance"):
            global_rules = guidance.get("global")
            if not isinstance(global_rules, list) or not global_rules:
                self.error("guidance.global", "must contain at least one rule")
            else:
                for index, rule in enumerate(global_rules):
                    self.require_string(rule, f"guidance.global[{index}]")
            levels = guidance.get("levels")
            if self.require_object(levels, "guidance.levels"):
                if set(levels) != LEVELS:
                    self.error("guidance.levels", "must define classic, refined, bold and avant-garde exactly once")
                for level_name in LEVELS:
                    level = levels.get(level_name)
                    path = f"guidance.levels.{level_name}"
                    if self.require_object(level, path):
                        self.require_string(level.get("direction"), f"{path}.direction")
                        if level.get("previewLayout") not in LAYOUTS:
                            self.error(f"{path}.previewLayout", f"must be one of {sorted(LAYOUTS)}")
                        rules = level.get("rules")
                        if not isinstance(rules, list) or not rules:
                            self.error(f"{path}.rules", "must contain at least one rule")
                        else:
                            for index, rule in enumerate(rules):
                                self.require_string(rule, f"{path}.rules[{index}]")

        accessibility = pack.get("accessibility")
        if self.require_object(accessibility, "accessibility"):
            invariants = accessibility.get("invariants")
            if not isinstance(invariants, list) or len(invariants) < 4:
                self.error("accessibility.invariants", "must declare at least four accessibility invariants")
            else:
                for index, invariant in enumerate(invariants):
                    self.require_string(invariant, f"accessibility.invariants[{index}]")

        examples = pack.get("examples")
        if not isinstance(examples, list) or not examples:
            self.error("examples", "must contain at least one format example")
        else:
            for index, example in enumerate(examples):
                path = f"examples[{index}]"
                if self.require_object(example, path):
                    format_name = example.get("format")
                    if format_name not in supports:
                        self.error(f"{path}.format", "must also appear in supports")
                    for field in ("title", "lead", "action"):
                        self.localized(example.get(field), f"{path}.{field}")

        assets = pack.get("assets", [])
        if not isinstance(assets, list):
            self.error("assets", "must be an array when present")
        else:
            for index, asset in enumerate(assets):
                path = f"assets[{index}]"
                if self.require_object(asset, path):
                    for field in ("path", "license", "attribution"):
                        self.require_string(asset.get(field), f"{path}.{field}")


def main(argv):
    if len(argv) != 2:
        print("usage: python3 tools/validate-style-pack.py <style-pack.json>", file=sys.stderr)
        return 2
    pack_path = Path(argv[1])
    try:
        pack = json.loads(pack_path.read_text(encoding="utf-8"))
    except OSError as error:
        print(f"error: cannot read {pack_path}: {error}", file=sys.stderr)
        return 2
    except json.JSONDecodeError as error:
        print(f"error: invalid JSON at line {error.lineno}, column {error.colno}: {error.msg}", file=sys.stderr)
        return 1

    validation = Validation()
    validation.validate(pack)
    if validation.errors:
        print(f"{pack_path}: invalid style pack", file=sys.stderr)
        for error in validation.errors:
            print(f"- {error}", file=sys.stderr)
        return 1
    print(f"{pack_path}: valid dumps Style Pack v1 ({pack['id']})")
    return 0


if __name__ == "__main__":
    raise SystemExit(main(sys.argv))
