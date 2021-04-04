__kernel void blur(
  __write_only image2d_t image,
  const float posx,
  const float posy,
  const float scale
) {
  int id = get_global_id(0);
  int idx = id % Width;
  int idy = (id / Width);

  float SSAmt = 8;

  float x0 = ((float)idx)/(float)Width;
  float y0 = ((float)idy)/(float)Height;
  //UV is now [-.5,.5] and [-.5,.5]
  float3 sumCol = (float3)(0,0,0);//(float4)(uv.x,uv.y,0.0,1);

  for (float offY = -SSAmt/2; offY<SSAmt/2; offY++){
    for (float offX = -SSAmt/2; offX<SSAmt/2; offX++){
      float x = x0+(offX/SSAmt)/(float)Width;
      float y = y0+(offY/SSAmt)/(float)Height;
      if (idx==1000&&idy==600){printf("(%f,%f)",x,y);}
      float2 uv = (float2)(x,y)-(float2)(0.5);
  

      float3 col = (float3)(0,0,.4);//(float4)(uv.x,uv.y,0.0,1);
      x = uv.x*scale+posx; //scaled x coordinate of pixel (scaled to lie in the Mandelbrot X scale (-2.5, 1))
      y = uv.y*scale+posy; //scaled y coordinate of pixel (scaled to lie in the Mandelbrot Y scale (-1, 1))

      float zx = x; // zx represents the real part of z
      float zy = y; // zy represents the imaginary part of z 

      int iteration = 0;
      int max_iteration = 64;

      //Do the iterationing
      while (zx*zx + zy*zy < 10000 && iteration < max_iteration){
        //float xtemp = zx*zx - zy*zy + x;
        float zsqx = (zx*zx)+ (zy*zy*-1);
        float zsqy = 2*zx*zy;
        
        zy = zsqx+x;
        zx = zsqy+y;
        iteration++;
        float amt=(((float)iteration)/(float)max_iteration);
        if (iteration == max_iteration){ // Belongs to the set

          col=(float3)(1,1,1);
          break;
        } else {
          col=mix(col,(0.5,.4,1.)*.8,pow(amt,(float)1.)*1.2);
        }
      }
      sumCol+=col;
    }
  }

  sumCol/=(float)(SSAmt*SSAmt);

  write_imagef(image, (int2)(idx,idy), (float4)(sumCol,1));
}
