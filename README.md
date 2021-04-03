# Fancy Math things
![Most Recent](https://github.com/cowsed/PrettyMath/blob/main/Gallery/2.png?raw=true)
![Made a while ago](https://github.com/cowsed/PrettyMath/blob/main/image.png?raw=true)

## Opencl workspace
Opencl pipeline 
feature list coming soon

## 2D attractors 
in the form of

```
xnew=f(x,y,a,b,c,d)
ynew=g(x,y,a,b,c,d)
plot(xnew,ynew)
x=xnew
y=ynew
```

Features:
- fine control over output parameters
- moderately functioning parser for math expressions (mostly a mess right now)
- Gradient Support! (I really like how this turned out)

Sources:
- AllenDang's giu and imgui libraries
- https://softologyblog.wordpress.com/2017/03/04/2d-strange-attractors/

Todo:
- Fix Expression Parser to correctly parse single variable functions(also optimize it so it makes bytecode instead of traversing a tree)(also maybe make a jit inside to generate a function and return a pointer to it then just call the function  when necessary(Thats a big stretch though, Id have to learn a lot of things))
- Implement ideas in Workspaces/ideas.txt
- Fix Plotting workspace so that it only makes cal
