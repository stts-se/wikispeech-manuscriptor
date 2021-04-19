Eftersom WikiExtractor.py inte lyckas expandera alla templates riktigt, och inte vet om att den gör fel, är det bättre att skriva ut templates som dom är och sen ta bort meningar som innehåller sådana.

Den här versionen av WikiExtractor.py gör det, med argumentet --no_templates.

Exempel:
```
python3 WikiExtractor.py solna.xml -o solna --no_templates
```

Problem:

Rader som "{{kolumner-slut}}" blir kvar i texten. Det är en del av uppmärkning för en tabell, inte ett template.


