# Fancy Math things

## Model Viewer
Can load arbitrary obj files.
Can generate a solid of revolution and calculate the volume. (Thanks BC Calc)
Can generate a solid where the side length is determined by a function and cross sections perpendicular to the x-axis

![Modelviewer Example Normals](https://github.com/cowsed/PrettyMath/blob/main/Gallery/ModelsNormal.png?raw=true)
![Modelviewer Example Shaded](https://github.com/cowsed/PrettyMath/blob/main/Gallery/ModelsShaded.png?raw=true)

## OpenGL Shader Editor
Uniform editors
single pass shader editor much like Shadertoy
supports setting int, float, and vec3 uniforms
renders to framebuffer and texture to be used in gui as a texture
![OpenGL Example](https://github.com/cowsed/PrettyMath/blob/main/Gallery/OpenGLExample1.png?raw=true)

## OpenCL workspace
Opencl pipeline  -  currently broken

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



Todo:
Remake Plotting workspace so its actually useful