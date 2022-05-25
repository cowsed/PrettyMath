# Pretty Math
A selection of math visualizations/art-ish tools

Uses Dear Imgui for gui

## Model Viewer
Can load arbitrary obj files.
Can generate a solid of revolution and calculate the volume. 
Can generate a solid of Extrusion where the side length is determined by a function and cross sections perpendicular to the x-axis and calculate volume


![Modelviewer Example Normals](https://github.com/cowsed/PrettyMath/blob/main/Gallery/ModelsNormal.png?raw=true)
Normal Shaded generated solids
![Modelviewer Example Shaded](https://github.com/cowsed/PrettyMath/blob/main/Gallery/ModelsShaded.png?raw=true)
Phong Shaded generated solids

## OpenGL Shader Editor
Uniform editors
single pass shader editor much like Shadertoy
supports setting int, float, and vec3 uniforms
renders to framebuffer and texture to be used in gui as a texture
![OpenGL Example](https://github.com/cowsed/PrettyMath/blob/main/Gallery/OpenGLExample1.png?raw=true)


## 2D attractors 
in the form of

```
xnew=f(x,y,a,b,c,d)
ynew=g(x,y,a,b,c,d)
plot(xnew,ynew)
x=xnew
y=ynew
```
![Example Attractor 1](https://github.com/cowsed/PrettyMath/blob/main/Gallery/2.png?raw=true)
![Example Attractor 2](https://github.com/cowsed/PrettyMath/blob/main/image.png?raw=true)

Features:
- fine control over output parameters
- [Expression parser](https://github.com/cowsed/Parser) to take new formulas at run time. Accepts ln, sin, cos, and more coming soon 
- Gradient Support and editor



