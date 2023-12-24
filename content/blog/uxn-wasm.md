---
title: "A WebAssembly Core for Uxn"
date: 2023-12-24
featured: true
commentURL: https://mas.to/@remko/111636802421264317
scripts:
- blog/uxn-wasm.js
styles:
- blog/uxn-wasm.css
---
While watching [a Strange Loop talk on concatenative
programming](https://www.youtube.com/watch?v=umSuLpjFUf8 "'Concatenative programming and stack-based languages' -- Douglas Creager, Strange Loop 2023"),
I learned about [Uxn](https://100r.co/site/uxn.html), 
a small virtual machine that runs games, editors, drawing programs, ... Uxn has been [ported to various platforms](https://github.com/hundredrabbits/awesome-uxn#emulators "Uxn Emulators"), including classic consoles such as the Nintendo DS, GBA, Playdate, ...

There also is a JavaScript version: [Uxn5](https://rabbits.srht.site/uxn5/ "Uxn5: Uxn/Varvara JavaScript system"). Since the Uxn5 core was a straight translation of the reference implementation to pure JavaScript, it wasn't very performant. After [my latest expeditions into WebAssembly](/blog/waforth "WAForth: Forth Interpreter+Compiler for WebAssembly"), I wanted to give writing a pure emulator in WebAssembly a try as well, so I created [uxn.wasm](https://github.com/remko/uxn.wasm "uxn.wasm"), a WebAssembly core for the Uxn virtual machine. 

<!--more-->
<div class="demos"><p>Here is uxn.wasm in action on some Uxn ROMs:</p><div class="uxn-roms"><figure class="uxn-rom"><div class="container" style="width:192px;height:192px;"><a target="_blank" rel="noreferrer" data-rom-url=1 href="https://rabbits.srht.site/uxn5/#r=YqACz4AIN6AC_4AKN6ACT4AMN6AAwCaAIjeAJDegAXeAIDegAZ2AkDegAgiAgDegBwNgAZGgBwRgAYugBQRgAYWgBwVgAX-gBgVgAXmAIjaAAT-gAEA5JoADMaAAfjiABzGAJIYUAAWCFB0JMQCAlhaAAAiAAQ0AoAAAgAExgAAQAQaAABGAAxyDFyFgAJwAoAPRgCw3gAswgCg3gA0wgCo3gECALxeAkjYmgAsxgBQFlDYmgA0xgBgAQoFSAQkYgB-BCYBcC4CSNoCUNoADYAGUDIEMBgMwOYABHwOAFAAFggoPYADBYAECAICDFoAgCSAAFIH0PTagAZwoIAAEIqABnIAgN4CCFoAICSAACaAEWaAEAGABeACgFFmgEABgAW6gQACPoEAABs9gABwBiiD_9iJCgAYB6yKAIQAEgiRTWUAAoKZgAGYFBWAAXIAACCAAFQaAAgsgAAsGgAMKIAAEAkAAEQIibAaAAwggAAMCImwCQAAAL4ABb2AADaAQADgVgAEwIYABMWyAPxyAAASAYD8FggkFOKAEWThsgSkm_-MVbGD_3hRsgA8zwACgCACAAIc4oAL5ODSgAAAFGA8YT2D_3w9YgKM15SJPbP__AP8B__8AAQD_AQABAQGAAzCAKDeACTCgAAg4gCo3gAGAJheAATBgADmAAIAmF6BAgVACgAUwgB2A8oEODgMwOIAoNwbPYP-DAYAuF4D6AuciQoBkD9EibARgAAAGgAQfYAADgA-BuCQwP6AD2TiALDegAy8XbA8mzwEBMCsgACImz4AGGDAqIAAYJCbPgBEEECbPgASBEQYGIiJCgAFsgAUaAGwnOCSvgABvFSGqIP_2IiJsJC8nOCSU7xVhgBAO9yIiYmyAwODw-OAQAHyCgAADfAAwEIAACDgAfIICfICA_oAHEBwCgnwAIkKC_gICAgD-gIB8gA8DfIKA_IAvBf6CBAgIEIE_gAKARwB-gCcKfIKC_oKCggD8goKAAoAvA4CAgnyADwCCgA8D_oCA8IBfggcBgAA"><img src="/blog/uxn-wasm/life.png" /></a></div><figcaption><a href="https://git.sr.ht/~rabbits/uxn/tree/main/item/projects/examples/demos/life.tal">Game of life</a></figcaption></figure><figure class="uxn-rom"><div class="container" style="width:288px;height:224px;"><a target="_blank" rel="noreferrer" data-rom-url=1 href="https://rabbits.srht.site/uxn5/#r=YKABFoAGN6ABIIAiN6AA4IAkN2AMWgAARG9uc29sCkEgQ2FyZCBHYW1lCkJ5IEh1bmRyZWQgUmFiYml0cwoxOCBEZWMgMjAyMwACgyc-QRcFgIIWgAGJAyAABWAClQIAgECACwkXgAAQgAAIIAAOgAgLARmAABFgBogCAICAgCkAFYEUBwogAAyAABABgRsrbAIAAoBAYAoSAIBCgJYWgAAJGGAKBYCUNoBGMDmAAz8DgBEJIAA5gJI2gESCEBEGgAkKIAAIgABgAm9AAB4GgBGBDgkBYAJgQAAPBoAYgR0HAmACUUAAAAKBUxAIIAAJgACAlhdgAesAAACAPoKjAB-ACB4BGQaAPhEgAANgAjCAURCANgsgAAagEYBgBoWATDCKgcwIB4GATRFgB-qLgdoBCgaA2IAQA9kigE6GIgRPEWAHx4gigBABtiKFbhcBAICCFoABiQMgAAqAPRCAAWADoQIAgAKAEAcFYASXAgCACIAcDQhgCn5gASUCAIAEH4AEgC4ADYEuBRmAAGADb4QlBAuAPRABgBEIXQIAAoBAYAjQiGgElhaAAguA1AEEQIALgO0AMoHdGQggACaAlDaARjA5gAM_oAACKSAAGICSNoBEgxAFFTmgAAUqgToABYIoCKAAMDg5oABIK4MmCqAA4CugAQEpIAAahRECODsDgHMLAQhgAs2AAICWF4BCgXcGCRhgCDegAIERDUMRAKAPf4AIN6AFf4AKgQUbDDdgCZKAIjaAAT8moABwOYBEMaAAcDiASDGAJIMUPUg5gEYxoABIOIBKMYAJgEwRgAqAThGAC4BREUAAAKABUICAN6ABoICQN6ACCYAgN2AI1GADtIBAQAe_oAJ5gBwBAuKBHAkKgCA3oBUVgEwxgI8KTjGAAIBQEYAAgFGABABSgAljUxFgCKdgAGRgAHVgCI6gEZiAABCAAAQmODg0YAR3QAAQBoAAEAkgAAICbIAAEUADsKAEAGAADgeAORgRAYog__MiQAMsgAIQIAADgDZsgAMQgAIQgAEZgAIRoDYABoADGIEQBIMoBmygNgCGgAOCNwb2IoA2gAIRgBMOBmAIYxiANpsaGQeAAxgQgQQBJA-AKABPhC0B3CKCJwA7nScAbIB_BgMYEYACEAGAXwMGgDYLgr0DURABBoD4PAEQioABDQQCgAERgAAHoBG8OBQGgAAIIAAMBoABCCAASgYgAGJsAoBTEIABCSAACKAQSGADbQJsgEwQgBWCEBRgYANcAmyAAASgEfM4FIBMEBiAFYuCVA1MEYABgFMRoBAwQAM4AoUiCE4RgBaAUBGAAIEbBHhAAxwCgQuGIAYQIAASYABjgGWAjBdUoBCQQAL2gYBQEAogACIGgE4QCyAANQaARAVOEBlgADmEKQkqoBDYQALMgAAGgGUEUBFgAB6ERCEPoBDAQAKxgFARoBCoQAKooBDwQAKiBoCACyAAAgJsBoBMgFQAB4CKAUwRgOABBBmDCA4gAAIibAQHIAAKBoA9EAmBDgNgAy0HglYZBoA5GBBg_qMGYAAQYAPHgAMcgD0RgAFgA22AVAQ5GBCANoAyCwJsgDYEgDkYEWACa4TXGRNgABGABAkgAAqAMIA-EYAAgFIRbMAAoAQAhjgIAUEBiiD_7yJPiRgDCCAAFIJcCQAEoBG8OBSAAguEKyXcIk9sgAIQYP-1oDIAKSAAA4ABbIAAEIAAiQMgABJg_7eAAAiAUoCBB6AAACkDbIABgBiDEgMDbIACgQ4CYP-PgQ4AAoJLAGmDSwVA_LyATBCBCAazgFEQgDYLgRQMp2D_jCAABqARIEABTommAAiBygJg_XSAxj7oIoABgFIRYPzGoBEIQAEkYAKXYAHkgAFgAkBAAVSgAyCgAAAmoAAguzo5gDA_gEQwOKAAEDmAKDcmgDU_gEaADktAOYAqNyagI344FIBAGYAABIBAP6AnzjiALDeAgYAvFyGqIP-8IiKARDCgAGg4gEYwoAAIOYABEIACYAR7gEQwgEYwJ6AAKDgnoAB4gF0HOKARq4ACgACA_AkYYAPCgCg2oAAsgHMCoBV8gFQAAYIbBxqALxcnoABYiDgAsII4AAGAOACJgTgAPIw4AAGEOACYiHEAt4JxAAKAcQBQknEAAoFxAyIibC-AtAEoN4HJA4g4gCqAkRQmF6AYAIAAB-84FIAABIAwP6AVzDiArC8CgC8XAYog_-MigACAJhdibKAEAI_PgAAEoAA4OoBEMDigADCARjA4T4A5GBBgAAeAKwPfImwPgFYCKDdPgSgGNjqgF-Q4L4JmADaEZgFAGYBpBEA_oCfOgWkhgYAvFwaABpsaGYAFCSAAEoAotqAAMDkFN4AqtqAACDgFN4CHAMKEh4fEACCAxAKgFcyC6gYmF6AbAIABg7cA9oK3gewAfIYngMUGgEQwoAAUOIKjAjg6OIBOhDYA4oDCAESA2YFeFTmAPRAYEIAABCY4oBT-ODTPBhhgAeiDQAGAPYAZhUKHaQqEgCw3T4ADGoAvF4GvEEYwJyegABA4oBGegAJgAacngLuBEQeATRCAAWACK4AHCwAEoBJCOBSABWABAoGjABCFVwAQgukAlIDsBlMQgAENJCKA9AMDgC8Xgc0BODiHYwChgGMAQ4djBk8QgANgAceAB4BjB1g4FIAPYACegcYASIpjAYBQgocFKjgUYAEliWMAcIhjBaSAAmAA34fHAVEQgNkAY4AHgMcLZDgUgApgADpg-74PgcsAqIvLBRfcgCw3z4S_ALiAv4K9CadPgAMaQACDDw-AjjIoN4AFgCYXoBWcgCw3oAYAgAKALxcBiiD_9iKAAIAmF4AoNqAAMDmAKDegFYyALDdPgACBVAUoNiGAKDeAKAfwIkJsD4A_MIDmAUEwgHoBFWyA1wBAgNcFkjYmgD8xgJkFlDYmgEExgR4AdIFHCC8XbA8kgCo3JIC3BwGAJheUYAAUg7MBIZSBV4GBJGwGgEAKB4BbC6ABASkgAA6AQRmAAASAMD-gFiQ4bAaAYAoHgHuFHABhhRwA9IAcBC8KB4A6hTkAMIQ5ARXUgDkiXwkHgAAJoAAAKCAABQKgFcxsAqAVlGwPD4AqN4AoN8-AChuGLwOALDdHgLQOgCg2oAAIOIAoN0-ACpsah1GE1QOgAAAmgPEhKjeAgIAuF2yAAIDGFoAAgMUWgGA_PoAAgMQWgMA_PoAlM4ILCAQ_gACAwxYmOIAbAMKCJAjANoCgPz6ADzOCTQdQPz4mgAM_PoFaIe0zJoABPz4-gPGzAmygDWSAqDegAAKAqjegDVKArDegDViBBQpegKw3gKI2HYABDYCSAoAIN4AFAAqBBTcMN2wudGhlbWUAYPYWAGD_XKANa4CAN6AQFIA4N4DMgD4X4DV2oA8FoA8FhQTvNQJhYQGKIP_0IoAFH-siYmAADYAAoA4AFaAN4YAgN2ygFACgFACFgABgAHcCiCkRoDV2JqAAyDgktIABYABbISGqgEUNImygNXagAMgHgBUSOjiCIAMmYABQgR4d9iIigACBgPsTgAUIDACAIjaAAT-gAFA5oA9UNYAkhg1CSTWgM96ALDegD4-AIDcAD4AABKAAFDoFgAAEOKAz5jgvVWyvNCaAAGD_5KAOmCeAARlgAE8nAQFgAEknoAEAOWAAQYAHEzhgADmgDpg5BiAADiImgAFg_7VvgLEVNWwjYP6DJLs6OYD-HKAOmDg0QP_gAIsAGyZgACggACNn4ADIeaA1diHvKiAADibvNCggAA2AEwtA_-liJzUhIWxiImyMn1oUbGD-Iyagf_88gDw3BqAA_zyAOjcagD8cgAwYgD8XbAeLIABmGQaAYwogAGSAF5uPGhkGgAcKIABZgAgHGYA-Ewdg_72APxOAAAQmOKA1djigAMiAAE86OLQvgL8PNE9gABagAAA4gCo3T2AAC4EKBCg3gC8XgokAAIGJgAcLOGwCwP9AAALABEECjUw5NK9A_7WAAIABGQaA-BMGgPALIAARgP8HGYAAJoAINyaACjeADDeBoAAAhWD_TAEGgGQLIP_0IgKgBYAQADuFECAgAAagD9yAIDcAoA-QFIACGQagD5AVBoAfCiAAIgaAAR-JTxegZACABQeAAGD--QGKIP_zIiAAA2DzbQCUABIjMzkAKDYlMi8AJQA0MzgtMzJAhRcIKikpMAA3LScvghKFLwU7JTc4KSiFMAA_hS8CMzkygBYFNywtKTAogkcCGTs_gWMBMimBMYUagHcPJjAzJy8pKAA4LCkAJTg4JYBoAyMzOTaDQAUAJjYzLymBXYKnBTc5NjotOoMwBiYlODgwKUCBvw4tKShAAEIAHjY9ACUrJS2AwoDXBzYlMgAlOyU9h76B7wgnJTIyMzgANjmHHh0PJTc9ABczKCkAQQAdKTApJzgAJyU2KEAYMzYxJTCLFAYAAAASJTYokC86IC0nODM2PQBBABYpJTopACg5MispMzJAETgRUBFoSFAAU1AAWFAAUlVOAEVhc3kATm9ybWFsAEhhcmSHpQMAAAABiAAAAogAAAOIAA8EBAULAgMEBQYHCAkKCwsLiQwAEYYZAQ0PiQwUFRUAADEyMzQ1Njc4OVhKAFEASwBBhHkCAgMEgE4cERMUFhgaHB8hIyUnKQACBQkQFhocICMmKQABAgKAeEAGBgcICAkQEBESEhMUFBUWFhcYGRkaGhsbHBwdHR4eISIiIyMkJCUlJiYnJygoKSlXaGl0ZV9NYWdlXzExAFJlZIULCUVtcHJlc3NfMTeICiNTbWFsbF9Qb3Rpb25fMgBCdWNrbGVyXzIAU2xpbWVfMgBSYXSADYgmADOFJgYzAFR1bm5lggoLQmF0XzMATWVkaXVthFECNABLgI8ONABGaWVuZF80AEltcF80iyQANYIkBTUARHJha4AHBEdvYmxpgBcETGFyZ2WEnQ42AEhlYXRlcl82AFNwZWOCCQRPcmNfNoonADeEJxU3AEdob3N0XzcAT2dyZV83AFN1cGVyhOwmOABUb3dlcl9TaGllbGRfOABFbGVtZW50YWxfOABCZWhvbGRlcl84ijQAOYo0CDkAV2l0Y2hfOYD2BHVzYV85imMBMTCKZAoxMABGYW1pbGlhcoAbAkRlbYIkEFdoaXRlX01hZ2VfMTEAUmVkhQsGQ29uc29ydIsKli8HUXVlZW5fMTOGCJZbCVJlZ25hbnRfMTWKCgJkX0SAdQlsXzIxAEJsYWNrhw9wXwASnBLMEvMTHRNCE2oTkhO5E-4UHRRSFIIUrhKqEtsTAhMtE1ITeROhE8gT_RQtFGAUkBS8ErYS5RMMEzQTWROCE6oT1xQMFD0UbBScFMgSwRLtExcTPBNhE4wTshPjFBQUSRR3FKUU0xTeFOwU_P-DADGAwODw-OAQAEBgOB4eOGBAABgYPH5mwwAAABg8PBgAAAgqHH8cKggAP0CAgICAQD__AIEAAP-UBwf8AgEBAQEC_IImAwAAOMaBAAM4jHgYgAA0fjxOhg4cOHL-gH4MEHwOjnwOHBw8bP8MHsb4QMD8Ds58OsTAwPzOznyAfj4MGDBgwIB8zs6AAoEHHRgwYHgcPDY-ZmbD7nNjbmNjY948ZszAwMDmfO5zY4AACN7-ZmB4YGNmfIEHE2BgYDxmxsDexmY8xsbGzv7mxsYwgYYSGAweDAwMDBw4YMZsbHh4bGzG4IAsBmZ-eMbu_taCsgfm9t7OxsZ8zoELCnz8ZmZmbGBgwHzmgNAbzn_uc2ZsZmZjwzpmYDwGBmZc_jBgwMDCxnzmZoAAAm4_w4EHEjwYw8PD08vfd2LDw2Y8PGbDw8OAFCYYGBh-xgwY_mDD_gAAPmZmZjsA4GBgfGZmfAAAADxmYGY8AAwGBj6AF4APHn5gPgAOGBgYPhgYAAAGfMzMeMJ8wGBgbHZmZgAwABiAyQIADACA0AkMOMBgZmx4bOYAgucQDAAAwGZ-fmtjAADAfGZmZmOAXwNuZnY8gg8ofGDgAAA7ZmY-BgcAwHx2YGBgAAAAPmA8BnwAADB-MDAwNhwAAOZmZm6AhwDmgM4QAAAA42t_PjYAAANmPBg8ZsCBHxI-zHgAAH4MfjF-AAwMGBgAMDAAgQAgMDAAPGbb28PbWjw8RtvH29tGPIGCgoKCg4VChoaGh4VSgQUi6OTl6YeF7ODh7YeF-PX2-4eFhvn6hoeF_P3-_4eJ7uvq74uJNQBGgTsAhoIFAEKCCwBChhGAWQSJioqKioo1AEeUNYNBkjUASI5rAEKeNQBJiaGCZYKhhaePoQBKlDWDO5LXAEubNQBCkdcATJprgOOP1wBNg9eRL5U1AE6YZZVrAE-B3SHo5OXph4Xs4OHth4X4jo_7h4WGwsOGh4X8wMf_h4nunp7vitcAUI41AYyNhzUAwYA1AcTFizUAUY5rAZydgGsB-fqAawG8vYBrAb6_hWsAQ4HXAFKC1w7w84aHhfDx8vOHhfT19veFNQH9_oChAevqizUARoI1gjsAQ4ILAEOGEYD7BImKioqKimsAR5Q1g0GSNQBIjmsAQ541AEmJoYJlgqGFp4-hAEqUNYM7ktcAS5s1AEOR1wBMmmuA44_XAE2D15EvlTUATphllWsAT4LXIPDzhoeF8PHy84eF9I6P94eFhsLDhoeF_MDH_4eJ7p6e74rXAFCONQGMjYc1AMGANQHExYs1AFGOawGcnYBrAfn6gGsBvL2AawG-v4VrAECB1wBSgtcg0NOGh4XQ0dLTh4XU1dbXh4WG2dqGh4Xc3d7fh4nm4uPnijUARoLXgwUAQIDjAECIEQeGhoeJioqKioprAEeVNYFBkzUASI9rgi-YNQBJiKGCL581AEqUNYNxktcAS5o1gOOP1wBMmzUAQJHXAE2D0ao1AE6YZZVrAE-C1yDQ04aHhdDR0tOHhdRWV9eHhYZeX4aHhdzOz9-Hieafn-eK1wBQjjUBWluHNQDJgDUBzM2LNQBRjmsBWFmAawFcXYBrAVRVgGsBiMiFawBBgdcAUoLXDdjbhoeF2NHS24eF1NXWgKEB2dqAoQHd3oChAeLjizUARoLXgt0AQYILAEGGEYD7BImKioqKimsAR5Q1g0GSNQBIjmsAQZ41AEmJoYJlgqGFp4-hAEqUNYM7ktcAS5s1AEGR1wBMmmuA44_XAE2PmwBAnWsATphllWsAT4LXINjbhoeF2NHS24eF1FZX14eFhsrLhoeF3M7P34eJ5p-f54rXAFCONQFaW4c1AMmANQHMzYs1AFGOawFYWYBrAVxdgGsBVFWAawGIyIVrAFOC0YcFCXh5enuHhXx9fn-JF4nXjjUJcHFyc4eFdHV2d481C5GSkpKSk5WWlpaWl6AFBpmampqamwCGAACYhQCHFIcfAamqhAGSHwCxhQCSPwSlpKWkpYEEkx-AAYEEkz8DlqWWpYEEsn-FvgCps7-FAJH_A6enpaeAIgOjpaOjkuCXFQ-4sbW5sbW4srW5sZC5sbWzjP8CugC6ggICtrG1gQuUHwCzgyiL_wmwsbu2sbuwALS2gQgCtrGQraILoaGloamxsamgpaCgkd8JpaSlqqamqqWkpZL-CKUApaoAAKqlAJIfAqilloIfApalqI-cCKyrq6urr66ur4AHAK2TvQOrr6-rmdgAq5oelf8RDyU3PQAAGDM2MSUwAAASJTYoskfAVTUA_4MAg_0AfosPhR8HAQcfPz8fBwGEL4Q3AcOBgP8JgcOA4Pj8_PjggJQ_AH6Sb4V_Bu_HgwEBq8eFBwXHxwEBAe-GB4SnBpMBAQGDx--FhoA3gA8BxzmBAADHhAcCc4fngAAAgYQHB8OxefHjx4kBhAcHf4Hz75PxcYOEBwfx4-PDkwDz4YQHBzkHvz8D8TGDhAcFxTs_PwMxhgcHf4HB8-fPnz-FBwKDMTGAAokHAufPn4QHBzw8mcPDmTw8hAcH4fPz8_Pjx5-EBwGDGYDAATGAhAcHOZOTh4eTkzmEBweH48PJwZmZPIQHAhGMnIAAACGEBzH4-Pjw4AAAAPj6_frlKhUKHx8fDwcAAAAfv1-vV6xUqAcT___n____Dx9_g-f-_37gyIIPCfD4_sHnf_9-BR2CHwAHgh8CfKC4gi8A4IIfAj4HGYcfB__-f-CY___zgx8j8_9__v__-Pj4__j5v-_4-Pj7-Pn__x8fH_8fn_33Hx8f3x-fgx8B-L6CHwD4gx8BH32BHwmfnz9AgICAgEA_hAcB_wCBAAD_hAcH_AIBAQEBAvyEBwX_AODw8OCGHwU_QJ-_v5-IPwL___-ABIU_A_n9_fmGPwU_QJi8vJiGb4JuCAAADAwYGAAwMIQOgoWADwYcN-p3_j86gwcKPj8uPzo_Pjc_Pz6BAQg_PGbb28PbWjyERwY8RtvH29tGhQ8DAC8AAoIHBQI_HwIKCoPdguaAqgb__PDgwIAAhAcHwAAABw8fHx-BBwocGBgDAADg8Pj4-IEHCjgYGP__Pw8HAwEAhAcHAIDA4PD8__-EBwofHx8PBwAAwBgYHIEHCvj4-PDgAAADGBg4gQcHAAEDBw8___-EB4YBhIeAkgH4-IR_gaICPx8fhn-EM4SHgkOEhwL4-PyB1YR_AR8fgmSEf4UvhIcEAAAkZnaIBwUPHz8_Pz-GB4NfhQcF8Pj8_Pz8hAcH59vnfhg8PlaCBwE4WIA7iAOI4YHliwAAAIMAAAWBCAH__4B7Ax8PAACDRwAfin2CPwP48AAAg0cg-P__f__n__9_Bxn_g-f__v__-___8____uCc_sHz_3__hR8GE_-D5_7__oAfAOeAHwfM_sHnf_9_BIJOAACAB4EIBgAPECAgICCGBwD_g6KBB4D_BAgEBAQEhgcFAABAICDAhMeAO4A_ACGDAITfB4FCJBgYJEKBgDuAPwCEgwAGAAAAApBBBIBkgQeAewEQD4JHAzA_Px-EeoJ9Af__gHsDCPAAAICDAAyH_wEFHYHfAPyE3wGgvIHfAT__gQAEAAAYABiBQIIHhg8H__78-PDgwICFBwZ_Px8PBwMBhAeE14TfB4DA4PD4_P7_hAcGAAD_58OBgYCkggcFz88AAPn5hgcBqqqAt4UHBwEDBw8fP3__hAcHyKEkUhS5UjiEBwX_AJmn5ZmAMYIHB72JmeWnmZG9hAcH__-_KVIFQACEBwMrBQgFgNCEBwPUoBCgiAcJPHb93_____88foAFBv___7___vuCB4EQBDw8PDx4gCcFPDw8_Pw4gKuArYIHgAiAPACAhA8DgAAAeIAxATw8hAeBDQAegj8LPD8fDAAAwPD4-Hw8ggcJ_Hw-Px8fBwAAAIAHCw8HAADA4PB4PB4PB4QHBwB4Oz8_Pz48ggcJPz4AAAMPHx8-PIIHAT8-h2WBeAd8_Pj44AAAAIAHgAgi-P369eoV6vX_-v361eoVCh9fv1-rVKtX_79fr1erVKj1-vyB0AIKBQOB1AKvXz-B4AJQoMCB5AT4___334ASgE4SwAAAH7-_r7vc7_f_X19fRyMQCIEYBP___77vgiMK__8fHx____999_-ECoFfggEC___3gQAGAAAoCAgoCIA9Av___4NfAPCBXwP____7gJIDHx8PB4CSgzoAUIM_AB-AFwgAAABfX19PRyKDfwT4-P--74EHAPuBfwQfH_9994EHAN-GQYJ_AyAAACCAjYV_gHcCAAAAgQcCYAAAhF-BZwQGAAD-_ICXAYAghAcLAAEBAAAAAQMAAAIBgQcKgIAAAACAwACAQICABwF_P4C3AQEAhAdCQAAAAQGBwPBAAACDEYPE8Acdf7_P____Bx8_w8_-_37guP795____-D4_MPnf_9-AgAAgICBAwcCAgDAgtEDB_jgwIFhggcBIACB7wP_-Pm9gO8C-_j9ge8D_x-fvYDvBd8fvx8HA4GhhAcA_4AAAvjgwIIHIfDA-Pn48eABAAH4_fj16BUoFR-fH48HgACAH78frxeoFKiBLwIfBwOCByIPA__-_f_-_v37AAMDAQEBAwf__7__f3-_7wBAwICAgMDQAIMAAgoFA4FthA8CUKDAhH0C6_36gYUCFwMFgY8D169_v4CWBejQgEAAwIEACgAAwMDiwOrS__8DgQAHAAADA0cDV0uBvYQEBPzw4MCAhxEIfx8PBwNX7x__gHUCqJDgge0C6vf4gQ8CFQkHhP2B-gCAg6iEUQEBAYQPA93r1emAfwPg1OjUgP0Du9erl4B_AwcrFyuGfyf-_Pjw4MCBAv369_3-_v37Bg0ZI0GBAwe_X--_f3-_72CwmMSCgcDQhZwAP4CdD4FA_7vX6_X6_P4EzPj0-v2Btir_z___fwcd_4PP_vv8__v-_-f___7gvP_B53_fP_rd61evXz9_JTMfr1-_gf8Kr9Xr8PwAAIDR6_WA_Qv4-Pj_-Pm_7_____uA3wsfHx__H5_99____9-B7wrr168PPwAAAJWr14T_AO-CGSL48PD6__r36leq1_35_fnVqVUpX_9f71fqVeu_n7-fq5WqlILxAPeBWQT_Hw8XAIwABjx-fn5-PAA"><img src="/blog/uxn-wasm/donsol.png"/></a></div><figcaption><a href="https://100r.co/site/donsol.html">Donsol</a>: A dungeon crawler card game</figcaption></figure><figure class="uxn-rom"><div class="container" style="width:350px;height:320px;"><a target="_blank" rel="noreferrer" data-rom-url=1 href="https://rabbits.srht.site/uxn5/#r=MqAs6YAIN6ABwIAKN6As5YAMN6ABZoAgN6ABXoAiN4AiNqAARjmgAAigBCyAQmACCaAABIENDjGAQmAB-4AiNoABP6AAUIIiCDeAQ2AB5qAAKIAwGwAAYAIJgASgBAEVAIAAsCEFMYDGFoACEAggAByBCQARgV0AK4BdFIAAMGAB3KAAAIAAMYCWFoABCCAADIEIOQogAAlAAAxgATdAAAZgAgpAAACgBlU0oAAAqCAACCZgAAchQP_0IiIAgDA_oAZXOLSABT-AKDehITSACBUqN2AA6rQnoAAEODQ4JzWhITQnoAAGgAsDISE1JoEWCYAPP6AAASggAB-AOgqgAAg4gCI2KyAAEIIeA6D__zqBPwE1tIEpAQAojxoAJoFUhUkAJIJ9gUsAJIBLABOCIINLBQY4NUAAMIChhFCPH4B-gDICAAQ4gjKD2AAhhNgVoBUmF6AGNYAsN4CFgC8XoAAmF2ygEYATAC2AEwAAhRMuBlU0JqD__yggAEomgDA_oAZXOIAAYAEBJzVgAPyAHxxgAPYnISE1gABgAO2AfxyC7IANAN-CDSoGODUiISagBlU1r6AAKKAACG9gADEibKABJhcPJIAqNySAKDeUgCAZgAAEgF4RBE04gCw3z4AvFyGUIP_mIk8CgpoCASYXhC0NoCcQuyYDgCoOOjmgA-iACgAfgAoBAGSAFRAUDjo5A4AKmwaACg4aGYAFDoHZCIAwH4AABKAEzYBcA4BBgC-C2AudgAAIIAA0JqAAATmD2wm0gAU_gCg3oSE0gAgFKjdg_ukigB6KwVT_byJsgAQMAAAAAID5EgaAQB8egPISgO4TgO0SgOkTgOgSBoDjE4YYHgQGgAEfHh4GgNYTbEZQUzoAQlVOUzoAQ0xJQ0sgVE8gQUREIEJVTk5JRVMhgE2BAwBggAAFAGAAZmZmghMdbP5sbP5sABg-YDwGfBgAAGZsGDBmRgA4bDhw3sx2gC-BOxQOHBgYGBwOAHA4GBgYOHAAAGY8_zyAQgQYGH4YGINeATAwgCwAfoVsExgYAAIGDBgwYEAAPGZudmZmPAAYgD8EGH4APGaAGBV-AH4MGAwGZjwADBw8bH4MDAB-YHwGgA8DPGBgfIAvAH6APwUwMAA8ZmaAAoAHAz4GDDiAyIACgHOBdoBlAjAYDIJ-An4AAIAMAhgwYIJnAQAYgH8Pam5gPgAYPGZmfmZmAHxmZoACFwA8ZmBgYGY8AHhsZmZmbHgAfmBgfGBgfoMHCmAAPmBgbmZmPgBmgTYDZgB4MIAAAngABoAAC2Y8AGZseHB4bGYAYIEAEH4Axu7-1sbGxgBmdn5-bmZmgLcAZoD3gG-AT4APAnZsNoF_AGyBHwFgPID_AX4YgQCAZwBmg28MZmY8GADGxsbW_u7GAIAMAjxmZoCHADyAJwF-BoDxAn4AeIF4AngA_4MACyRmZgAkJAA8QgAAfoAAGRgYPDwYAAAA_2ZCQmZ-QkIAAAYABQAAYAAQ"><img src="/blog/uxn-wasm/bunnymark.png"/></a></div><figcaption><a href="https://git.sr.ht/~rabbits/uxn/tree/main/item/projects/examples/demos/bunnymark.tal">BunnyMark</a> benchmark</figcaption></figure></div><p class="about">The above demos are embedded links to the <a href="https://rabbits.srht.site/uxn5/">Uxn5 demo page</a>, with the full (<a href="http://wiki.xxiivv.com/site/ulz_format">compressed</a>) ROM encoded in the link.</p></div>

## Implementation

The core [Uxn opcode set](https://wiki.xxiivv.com/site/uxntal_reference.html "Uxn opcode reference") consists of ±32 base opcodes, but each opcode has
several variants (depending on which of the 2 stacks they operate on, whether they
leave their operands on the stack, whether they work on 8- or 16-bit data, and
different combinations of these modes). In total, that makes 256 separate opcodes.

The [Uxn reference C
implementation](https://git.sr.ht/~rabbits/uxn/tree/main/item/src/uxn.c) has an
implementation of only the base opcodes, and implements the variant-specific
logic at runtime. This gives very compact and readable code. For uxn.wasm, I
wanted to squeeze as much performance out of the execution as possible, so I
wanted specialized opcodes for each of the variants. However, because every
opcode has different variants that look very similar, I didn't want to
hand-write (and debug) 256 separate opcodes. I therefore wrote the WebAssembly
code of the base opcodes, mirroring the logic of the reference implemetation,
and used a custom macro-enabled WebAssembly to expand this into the different variants. 

For example, take the C reference implementation of the `ADD` opcode:

```c
t=T;
n=N;
SET(2,-1);
T = n + t;
```

This opcode is specified as follows in the macro language:

```wasm
(local.set $t (#T))
(local.set $n (#N))
(#set 2 -1)
(#T! (i32.add (local.get $n) (local.get $t)))
```

The `#T` and `#N` macros load the top and next item from the stack, the `#set`
macro does the necessary stack adjustments, and `#T!` stores the top of stack. 

[The assembly script](https://github.com/remko/uxn.wasm/blob/master/scripts/emuwasm.js "uxn.wasm opcode definitions and generator") then processes these specifications, expands the macros differently for each of the different opcode variants, and then does some small local optimizations on the result.

The base case of the above `ADD` opcode becomes:

```wasm
;; ADD
(local.set $t (i32.load8_u offset=0x10000 (local.get $wstp)))
(local.set $n (i32.load8_u offset=0x10000 
  (i32.add (local.get $wstp) (i32.const 1))))
(local.set $wstp (i32.add (local.get $wstp) (i32.const 1)))
(i32.store8 offset=0x10000 (local.get $wstp) 
  (i32.add (local.get $n) (local.get $t)))
```

The generated variant that uses the return stack is the same, except it uses a different
stack pointer and offset:

```wasm
;; ADDr
(local.set $t (i32.load8_u offset=0x10100 (local.get $rstp)))
(local.set $n (i32.load8_u offset=0x10100 
  (i32.add (local.get $rstp) (i32.const 1))))
(local.set $rstp (i32.add (local.get $rstp) (i32.const 1)))
(i32.store8 offset=0x10100 (local.get $rstp) 
  (i32.add (local.get $n) (local.get $t)))
```

The generated variant that leaves the operands on the stack only differs in the stack adjustments:

```wasm
;; ADDk
(local.set $t (i32.load8_u offset=0x10000 (local.get $wstp)))
(local.set $n (i32.load8_u offset=0x10000 
  (i32.add (local.get $wstp) (i32.const 1))))
(local.set $wstp (i32.add (local.get $wstp) (i32.const 255)))
(i32.store8 offset=0x10000 (local.get $wstp) 
  (i32.add (local.get $n) (local.get $t)))
(br $loop)
```

The final result of running the generator is a [WebAssembly module](https://github.com/remko/uxn.wasm/blob/master/src/uxn.wat "uxn.wasm WebAssembly module") consisting of a big switch (`br_table`) with one case of each of the 256 opcodes. The entire binary module is about 19kB large.


## Compliance Testing & Coverage

Apart from testing that the various ROMs run correctly on uxn.wasm, there also
is an 'official' [Uxn opcode test
suite](https://git.sr.ht/~rabbits/uxn-utils/tree/main/item/cli/opctest/opctest.tal) that
tests the various opcodes and their edge cases. These tests are run against uxn.wasm on a regular
basis in CI to make sure that the implementation stays compliant with the reference.

I also wanted to have an idea how much of the WebAssembly code these tests
covered, and which parts were still left uncovered (and so might need extra
tests).  There isn't really any tooling for testing WebAssembly coverage: normal
people don't write code directly in WebAssembly, so you're usually only
interested in coverage information at a higher level. 

To get test coverage information, I used
[`wasm2c`](https://github.com/WebAssembly/wabt/tree/main/wasm2c) to convert the
WebAssembly module into C code, compiled this C code into [a standalone
(native) Uxn interpreter](https://github.com/remko/uxn.wasm/tree/main/src/cli
"uxn.wasm CLI") with the C compiler's coverage instrumentation enabled, used it
to run the tests, and then used `llvm-cov` to annotate the source code with the
collected coverage data. Because the WebAssembly module is essentially just one
big switch of different cases, it's easy to map the C code with annotated
coverage information back to WebAssembly purely on sight.

For example, by looking at the annotated opcode switch, you can tell that all basic opcodes up until 64 (`JMI`) are covered, but then some subsequent variants are not:

```
Count   Source
---------------------------------
4.48k   switch (var_i0) {
...
    4     case 62: goto var_B197;
    5     case 63: goto var_B196;
   36     case 64: goto var_B195;
    0     case 65: goto var_B194;
    0     case 66: goto var_B193;
    0     case 67: goto var_B192;
```

Looking at the individual opcodes, it's also simple to see which parts aren't covered.
For example, this part of the `DIV` opcode showed that the test suite was missing
a division by zero test:

```
Count   Source
-----------------------------------------
    7   var_B232:;
...
    7   var_i1 = !(var_i1);
    7   if (var_i1) {
    0     var_i1 = 0u;
    7   } else {
    7     var_i1 = var_l4;
    7     var_i2 = var_l3;
    7     var_i1 = DIV_U(var_i1, var_i2);
    7   }
    7   i32_store8(&instance->w2c_memory, 
          (u64)(var_i0) + 65536, var_i1);
```

An extra bonus of adding the `wasm2c` build infrastructure is that you also get
a small (70kB) native Uxn implementation binary as a side-effect. It's still
about twice as big as the compiled 100-lines reference implementation, but this
is because there is specialzed code for each opcode, and `wasm2c` code is
slightly more verbose.


## Benchmarks

The initial goal was to get a performant Uxn system in the browser,
preferably comparable in speed to the native implementation. To get an idea of the
performance, I ran the following 2 benchmark programs (written in Uxntal):

- [`mandelbrot.tal`](https://codeberg.org/jan0sch/uxn/src/commit/a11660f57db0c6ec6fa1b1e048674d86e2a818fd/projects/examples/demos/mandelbrot.tal): draws a Mandelbrot fractal
- [`primes32.tal`](http://git.phial.org/d6/nxu/src/branch/main/primes32.tal): computes prime numbers

These programs were run against the following Uxn implementations:

- `uxn5`: The pure JavaScript implementation
- `uxncli`: The reference Uxn implementation, compiled to a native binary
- `uxn.wasm`: The WebAssembly implementation described here, running in different browsers: Safari 17.1, Chrome 119, Firefox 121.0, and Node.js 21.2.0.
- `uxn.wasm-cli`: The `uxn.wasm` WebAssembly module, translated to C using `wasm2c`, and compiled to a native binary.

Running these tests on my MacBook Air M2 gives the following run times:

|                         | `mandelbrot` | `primes32` |
|-------------------------|--------------|------------|
|`uxn5` (Safari)          |       46.42s |     42.40s |
|`uxn5` (Chrome)          |       44.61s |     26.64s |
|`uxn5` (Firefox)         |       61.72s |     56.16s |
|`uxn5` (Node.js)         |       37.49s |     23.35s |
|`uxncli`                 |        1.81s |      2.31s |
|**`uxn.wasm`** (Safari)  |        1.29s |      1.51s |
|**`uxn.wasm`** (Chrome)  |        1.59s |      1.91s |
|**`uxn.wasm`** (Firefox) |        1.39s |      1.71s |
|**`uxn.wasm`** (Node.js) |        1.57s |      1.90s |
|**`uxn.wasm-cli`**       |        1.81s |      2.29s |

As expected, `uxn.wasm` is at least an order of magnitude faster than the pure
JavaScript implementation.  It is even slightly faster than the natively
compiled pure reference implementation (which isn't written for speed, and
doesn't have specialized opcode variant code). The native `uxn.wasm` (compiled via
`wasm2c`) is slightly slower than the direct WebAssembly version.

A sidenote: initial versions of uxn.wasm were even faster. By using
a downward-growing stack, it was possible to read 16-bit words in one
WebAssembly opcode from the stack, without needing endianness swaps.
Unfortunately, I later realized that Uxn uses circular stacks, so reading
16-bit words isn't possible, and so several of these optimizations were no
longer possible.  Even though not many ROMs depend on circular stacks, I still
chose to go with 100% compliance at the cost of performance.


## Using Uxn.wasm in JavaScript

Even though uxn.wasm was designed as a drop-in core for Uxn5, uxn.wasm is also
packaged so you can use it in JavaScript without requiring Uxn5 (which isn't packaged).

The [uxn.wasm npm module](https://www.npmjs.com/package/uxn.wasm) ships with
extra utilities under the `util` submodule to easily run Uxn programs,
including a [Uxntal](https://wiki.xxiivv.com/site/uxntal.html "The Uxntal language") assembler
(`asm`), and utility devices (e.g. a `LogConsole` console device that logs
output to `console`).

The example below runs a Uxntal program to compute prime numbers below 65536, and writes
them to the console. 

```javascript
import { Uxn } from "uxn.wasm";
import { asm, mux, LogConsole } from "uxn.wasm/util";

(async () => {
  const uxn = new Uxn();

  // Initialize the system with 1 device: a console at device offset 0x10 that
  // logs output using `console.log`.
  await uxn.init(mux(uxn, { 0x10: new LogConsole() }));

  // Assemble the program written in Uxntal assembly language into a binary ROM 
  // using `asm`, and load it into the core.
  uxn.load(
    asm(`
( Source: https://git.sr.ht/~rabbits/uxn/tree/main/item/projects/examples/exercises )

|0100 ( -> ) @reset
  #0000 INC2k
  &loop
    DUP2 not-prime ?&skip
      DUP2 print/short #2018 DEO
      &skip
    INC2 NEQ2k ?&loop
  POP2 POP2
  ( flush ) #0a18 DEO
  ( halt ) #010f DEO
BRK

@not-prime ( number* -- flag )
  DUP2 ,&t STR2
  ( range ) #01 SFT2 #0002 LTH2k ?&fail
  &loop
    [ LIT2 &t $2 ] OVR2 ( mod2 ) DIV2k MUL2 SUB2 ORA ?&continue
      &fail POP2 POP2 #01 JMP2r &continue
    INC2 GTH2k ?&loop
  POP2 POP2 #00
JMP2r

@print ( short* -- )
  &short ( short* -- ) SWP print/byte
  &byte  ( byte   -- ) DUP #04 SFT print/char
  &char  ( char   -- ) #0f AND DUP #09 GTH #27 MUL ADD #30 ADD #18 DEO
JMP2r
`)
  );

  // Start running at the default offset (0x100)
  uxn.eval();
})();
```
