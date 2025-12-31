# Advent of Code 2025 - Day 1

CLI qui calcule une valeur finale à partir d'instructions L/R sur un accumulateur initial de 50 et compte combien de fois le cadran passe par la valeur 0 (y compris lorsqu'il s'y arrête).

## Day 2

Deuxième CLI qui lit des intervalles `min-max` séparés par des virgules, identifie toutes les valeurs composées d'une séquence de chiffres répétée au moins deux fois (11, 6464, 123123, 123123123, etc. sans zéros initiaux) et affiche la somme de ces identifiants « invalides » présents dans les intervalles.

## Day 3

Chaque ligne d'entrée représente une banque de batteries identifiée par une suite de chiffres. Pour chaque banque, allumez exactement 12 batteries (en conservant l'ordre) afin de former la tension la plus élevée possible, puis additionnez les meilleures tensions de toutes les banques et affichez la somme totale. Exemple :

```
987654321111111 -> 987654321111
811111111111119 -> 811111111119
234234234234278 -> 434234234278
818181911112111 -> 888911112111
Total           -> 3121910778619
```

## Day 4

Chaque ligne d'entrée est une grille de `.` et `@` représentant des rouleaux de papier. À chaque itération, on repère les rouleaux ayant strictement moins de 4 rouleaux dans leurs 8 cases adjacentes, on les retire, puis on recommence jusqu'à stabilisation. Le programme affiche le nombre total de rouleaux retirés.

## Day 5

Le fichier d'entrée contient une liste de plages `min-max` (inclusives) d'identifiants « frais » (chevauchements possibles). Un identifiant est frais s'il appartient à au moins une plage. Le programme affiche le nombre total d'identifiants distincts considérés comme frais (taille de l'union des plages).

## Day 6

Le fichier d'entrée représente une feuille d'exercices avec plusieurs problèmes côte à côte. Chaque nombre est encodé en colonne (lecture de haut en bas) et les colonnes se lisent de droite à gauche dans chaque problème. En bas du problème, l'opération (`+` ou `*`) s'applique à tous les nombres du problème. Les problèmes sont séparés par une colonne entièrement composée d'espaces. Le programme calcule chaque résultat puis affiche la somme de tous les résultats.

## Day 7

Le fichier d'entrée représente une pièce en grille contenant un point de départ `S` et des séparateurs `^`. Un laser est tiré depuis `S` et se déplace vers le bas. Quand le laser atteint un `^`, le temps se divise en deux : dans une timeline il repart depuis la colonne de gauche, dans l'autre depuis la colonne de droite (sur la ligne suivante). Le programme affiche le nombre total de timelines possibles.

## Day 8

Le fichier d'entrée contient des positions `X,Y,Z` de boîtes de jonction. On relie les boîtes par paires en partant des distances euclidiennes les plus courtes (sans répéter une paire) jusqu'à ce qu'il ne reste plus qu'un seul circuit. Le programme affiche le produit des coordonnées X des deux boîtes reliées lors de la connexion qui crée ce circuit unique.

## Day 9

Le fichier d'entrée contient une liste ordonnée de coordonnées `x,y` de tuiles rouges formant une boucle. Chaque tuile rouge est connectée à la précédente et à la suivante (et la liste « wrap ») par une ligne orthogonale de tuiles vertes ; toutes les tuiles à l'intérieur de la boucle sont aussi vertes. On cherche un rectangle dont deux coins opposés sont des tuiles rouges et dont toutes les tuiles couvertes sont rouges ou vertes ; le programme affiche la plus grande aire possible (en nombre de cases, bords inclus).

## Day 10

Le fichier d'entrée contient une machine par ligne : un motif dans `[...]` (ignoré), une liste de boutons dans `(...)` (chaque bouton incrémente de 1 certains compteurs), puis des objectifs de tension dans `{...}`. Les compteurs démarrent à 0 et l'on cherche le nombre minimal total d'appuis de boutons pour atteindre exactement les objectifs ; le programme affiche la somme de ces minima.

## Day 11

Le fichier d'entrée décrit un graphe orienté de serveurs sous la forme `device: out1 out2 ...`. Les données démarrent sur `svr` et s'arrêtent quand elles atteignent `out`. Le programme calcule le nombre total de chemins distincts de `svr` à `out` qui visitent aussi `dac` et `fft`.
