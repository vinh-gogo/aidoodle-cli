import os
for root, _, files in os.walk('internal'):
    for f in files:
        if f.endswith('.go'):
            for line in open(os.path.join(root,f), encoding='utf-8', errors='ignore'):
                if 'case "/' in line:
                    print(f'{root}/{f}: {line.strip()}')
