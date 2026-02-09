/*
 * Copyright 2025 Alby Hernández
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

const METADATA_REGEX = /<!--MAGEC_META:.*?:MAGEC_META-->\n?/gs;

export function stripMetadata(text) {
    if (!text) return text;
    return text.replace(METADATA_REGEX, '').trimStart();
}

export function extractMetadata(text) {
    if (!text) return null;
    const match = text.match(/<!--MAGEC_META:(.*?):MAGEC_META-->/s);
    if (!match) return null;
    try {
        return JSON.parse(match[1]);
    } catch {
        return null;
    }
}
