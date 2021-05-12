#!/usr/bin/env python3
#
# Copyright (C) 2021 The Android Open Source Project
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

from fontTools import ttLib
from fontTools.ttLib.tables import otTables
from nototools import font_data

import contextlib
import re
import sys

REF_CODE_POINT = 0x1F1E6
NEW_CODE_POINT = 0x10FF00
NEW_GLYPH_ID = 'u10FF00'

def main():
  in_file = sys.argv[1]
  out_file = sys.argv[2]

  with contextlib.closing(ttLib.TTFont(in_file, recalcTimestamp=False)) as ttf:
    cmap = ttf.getBestCmap()
    gsub = ttf['GSUB'].table

    # Obtain Version string
    m = re.search('^Version (\d*)\.(\d*)', font_data.font_version(ttf))
    if not m:
      print('The font does not have proper version string.')
      sys.exit(1)
    major = m.group(1)
    minor = m.group(2)
    # Replace the dot with space since NotoColorEmoji does not have glyph for dot.
    char_seq = '%s %s' % (major, minor)
    glyphs = tuple([cmap[ord(x)] for x in char_seq])

    # Update Glyph metrics
    ttf.getGlyphOrder().append(NEW_GLYPH_ID)
    refGlyphId = cmap[REF_CODE_POINT]
    ttf['hmtx'].metrics[NEW_GLYPH_ID] = ttf['hmtx'].metrics[refGlyphId]
    ttf['vmtx'].metrics[NEW_GLYPH_ID] = ttf['vmtx'].metrics[refGlyphId]

    # Add new Glyph to cmap
    font_data.add_to_cmap(ttf, { NEW_CODE_POINT : NEW_GLYPH_ID })

    # Add new feature to GSUB
    # Add lookup for version
    lookups = gsub.LookupList.Lookup
    new_lookup = otTables.Lookup()
    new_lookup.LookupType = 2  # Multiple Substitution Subtable.
    new_lookup.LookupFlag = 0
    new_subtable = otTables.MultipleSubst()
    new_subtable.mapping = { NEW_GLYPH_ID : glyphs }
    new_lookup.SubTable = [ new_subtable ]
    new_lookup_index = len(lookups)
    lookups.append(new_lookup)

    # Add feature
    feature = next(x for x in gsub.FeatureList.FeatureRecord if x.FeatureTag == "ccmp")
    feature.Feature.LookupListIndex.append(new_lookup_index)

    # save to ttf
    ttf.save(out_file)

if __name__ == '__main__':
  main()
