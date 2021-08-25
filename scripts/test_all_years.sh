for Year in {2006..2021}
do
    ./sec dd $Year
    for Month in {1..12}
    do
        echo trying $Year/$Month
        ./sec de $Year/$Month
        ./sec dest
    done
done